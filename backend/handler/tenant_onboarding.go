package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/internal/totp"
	"github.com/raufhm/whatsapp-testing/whatsapp"
)

type signupTenantRequest struct {
	TenantName    string `json:"tenant_name"`
	Name          string `json:"name"`
	AdminName     string `json:"admin_name"`
	AdminEmail    string `json:"admin_email"`
	Email         string `json:"email"`
	AdminWhatsapp string `json:"admin_whatsapp"`
	Whatsapp      string `json:"whatsapp"`
	WhatsappNumber string `json:"whatsapp_number"`
}

func (d *DashboardHandler) handleSignupTenant(w http.ResponseWriter, r *http.Request) {
	var req signupTenantRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	tenantName := strings.TrimSpace(req.TenantName)
	if tenantName == "" {
		tenantName = strings.TrimSpace(req.Name)
	}
	adminName := strings.TrimSpace(req.AdminName)
	if adminName == "" {
		adminName = strings.TrimSpace(req.Name)
	}
	adminEmail := strings.TrimSpace(req.AdminEmail)
	if adminEmail == "" {
		adminEmail = strings.TrimSpace(req.Email)
	}
	adminWhatsapp := strings.TrimSpace(req.AdminWhatsapp)
	if adminWhatsapp == "" {
		adminWhatsapp = strings.TrimSpace(req.Whatsapp)
	}
	if adminWhatsapp == "" {
		adminWhatsapp = strings.TrimSpace(req.WhatsappNumber)
	}

	if tenantName == "" || adminName == "" {
		writeAPIError(w, 400, "VALIDATION_ERROR", "tenant name and admin name are required")
		return
	}
	if adminEmail == "" && adminWhatsapp == "" {
		writeAPIError(w, 400, "VALIDATION_ERROR", "admin email or whatsapp number is required")
		return
	}

	tenant, op, rawToken, err := d.Auth.SignupTenant(tenantName, adminName, adminEmail, adminWhatsapp)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateWhatsapp) {
			writeAPIError(w, 409, "DUPLICATE_WHATSAPP", "this WhatsApp number is already registered")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"tenant":                      tenant,
		"tenant_id":                   tenant.ID,
		"operator":                    op,
		"operator_id":                 op.ID,
		"verification_token":          rawToken,
		"temp_token":                  rawToken,
		"setup_url":                   "/dashboard/verify-email?token=" + rawToken,
		"email_verification_required": op.Email != "",
	})
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (d *DashboardHandler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := DecodeJSONBody(r, &req); err != nil || strings.TrimSpace(req.Token) == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "verification token is required")
		return
	}

	tenant, op, setupToken, err := d.Auth.VerifyEmailToken(strings.TrimSpace(req.Token))
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			writeAPIError(w, 400, "INVALID_TOKEN", "verification token not found or expired")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":      "verified",
		"verified":    true,
		"tenant":      tenant,
		"tenant_id":   tenant.ID,
		"operator":    op,
		"operator_id": op.ID,
		"setup_token": setupToken,
		"temp_token":  setupToken,
		"setup_url":   "/dashboard/totp/setup/" + setupToken,
	})
}

func (d *DashboardHandler) handleTOTPSetup(w http.ResponseWriter, r *http.Request, token string) {
	if token == "" {
		writeAPIError(w, 400, "INVALID_TOKEN", "setup token is required")
		return
	}

	op, tenant, plainSecret, err := d.Auth.GetTOTPSetupInfo(token)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			writeAPIError(w, 404, "INVALID_TOKEN", "setup token not found or expired")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	accountLabel := op.Email
	if accountLabel == "" {
		accountLabel = op.WhatsappNumber
	}
	if accountLabel == "" {
		accountLabel = op.Name
	}

	otpauthURL := totp.GenerateOtpauthURL(accountLabel, plainSecret)
	qrCodeDataURL, err := totp.GenerateQRCodeDataURL(otpauthURL)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to generate qr code")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"secret":           plainSecret,
		"otpauth_url":      otpauthURL,
		"qr_code":          qrCodeDataURL,
		"qr_code_data_url": qrCodeDataURL,
		"operator_id":      op.ID,
		"tenant_id":        tenant.ID,
	})
}

type verifyTOTPSetupRequest struct {
	Token      string `json:"token"`
	SetupToken string `json:"setup_token"`
	Code       string `json:"code"`
	TOTPCode   string `json:"totp_code"`
}

func (d *DashboardHandler) handleTOTPVerifySetup(w http.ResponseWriter, r *http.Request) {
	var req verifyTOTPSetupRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		token = strings.TrimSpace(req.SetupToken)
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = strings.TrimSpace(req.TOTPCode)
	}

	if token == "" || code == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "token and totp code are required")
		return
	}

	op, backupCodes, session, err := d.Auth.VerifyTOTPSetup(token, code)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidTOTPCode) {
			writeAPIError(w, 400, "INVALID_CODE", "invalid totp code")
			return
		}
		if errors.Is(err, storage.ErrTokenNotFound) {
			writeAPIError(w, 400, "INVALID_TOKEN", "setup token not found or expired")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((8 * time.Hour).Seconds()),
	})

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":       "success",
		"verified":     true,
		"backup_codes": backupCodes,
		"user":         op,
		"session_id":   session.ID,
	})
}

func (d *DashboardHandler) handleAcceptInvitationInfo(w http.ResponseWriter, r *http.Request, token string) {
	if token == "" {
		writeAPIError(w, 400, "INVALID_TOKEN", "invitation token is required")
		return
	}

	inv, tenant, err := d.Auth.GetInvitationByToken(token)
	if err != nil {
		if errors.Is(err, storage.ErrInvitationNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "invitation not found or expired")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"invitation":       inv,
		"invitation_token": token,
		"tenant_name":      tenant.Name,
		"tenant_id":        tenant.ID,
		"role":             inv.Role,
		"whatsapp_number":  inv.WhatsappNumber,
		"email":            inv.Email,
	})
}

type signupOperatorRequest struct {
	Token           string `json:"token"`
	InvitationToken string `json:"invitation_token"`
	Name            string `json:"name"`
	WhatsappNumber  string `json:"whatsapp_number"`
	Whatsapp        string `json:"whatsapp"`
	Email           string `json:"email"`
}

func (d *DashboardHandler) handleSignupOperator(w http.ResponseWriter, r *http.Request) {
	var req signupOperatorRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		token = strings.TrimSpace(req.InvitationToken)
	}
	name := strings.TrimSpace(req.Name)

	if token == "" || name == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "token and name are required")
		return
	}

	whatsapp := strings.TrimSpace(req.WhatsappNumber)
	if whatsapp == "" {
		whatsapp = strings.TrimSpace(req.Whatsapp)
	}
	email := strings.TrimSpace(req.Email)

	op, setupToken, err := d.Auth.AcceptInvitationAndSignupOperator(token, name, whatsapp, email)
	if err != nil {
		if errors.Is(err, storage.ErrInvitationNotFound) {
			writeAPIError(w, 400, "INVALID_INVITATION", "invitation not found or expired")
			return
		}
		if errors.Is(err, storage.ErrDuplicateWhatsapp) {
			writeAPIError(w, 409, "DUPLICATE_WHATSAPP", "this WhatsApp number is already registered")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"status":      "created",
		"operator":    op,
		"operator_id": op.ID,
		"setup_token": setupToken,
		"temp_token":  setupToken,
		"setup_url":   "/dashboard/totp/setup/" + setupToken,
	})
}

type backupCodeLoginRequest struct {
	TenantID       string `json:"tenant_id"`
	Identifier     string `json:"identifier"`
	Email          string `json:"email"`
	WhatsappNumber string `json:"whatsapp_number"`
	BackupCode     string `json:"backup_code"`
	Code           string `json:"code"`
}

func (d *DashboardHandler) handleBackupCodeLogin(w http.ResponseWriter, r *http.Request) {
	var req backupCodeLoginRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	tenantStr := req.TenantID
	if tenantStr == "" {
		tenantStr = r.Header.Get("X-Tenant")
	}
	tenantID, err := uuid.Parse(tenantStr)
	if err != nil {
		writeAPIError(w, 400, "TENANT_REQUIRED", "tenant_id is required")
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.WhatsappNumber)
	}

	code := strings.TrimSpace(req.BackupCode)
	if code == "" {
		code = strings.TrimSpace(req.Code)
	}

	if identifier == "" || code == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "identifier and backup_code are required")
		return
	}

	op, session, remaining, err := d.Auth.VerifyBackupCodeAndLogin(tenantID, identifier, code)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidBackupCode) || errors.Is(err, storage.ErrOperatorNotFound) {
			writeAPIError(w, 401, "INVALID_CREDENTIALS", "invalid identifier or backup code")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((8 * time.Hour).Seconds()),
	})

	WriteJSON(w, http.StatusOK, map[string]any{
		"user":                   op,
		"session_id":             session.ID,
		"backup_codes_remaining": remaining,
	})
}

type recoveryRequest struct {
	TenantID       string `json:"tenant_id"`
	Identifier     string `json:"identifier"`
	Email          string `json:"email"`
	WhatsappNumber string `json:"whatsapp_number"`
}

func (d *DashboardHandler) handleRecoveryRequest(w http.ResponseWriter, r *http.Request) {
	var req recoveryRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	tenantStr := req.TenantID
	if tenantStr == "" {
		tenantStr = r.Header.Get("X-Tenant")
	}
	tenantID, err := uuid.Parse(tenantStr)
	if err != nil {
		writeAPIError(w, 400, "TENANT_REQUIRED", "tenant_id is required")
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.WhatsappNumber)
	}

	if identifier == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "identifier is required")
		return
	}

	token, err := d.Auth.RequestRecovery(tenantID, identifier)
	if err != nil {
		if errors.Is(err, storage.ErrOperatorNotFound) {
			// Return generic message to prevent user enumeration
			WriteJSON(w, http.StatusOK, map[string]any{
				"status":  "recovery_initiated",
				"message": "If the account exists, recovery instructions have been sent",
			})
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":         "recovery_initiated",
		"message":        "If the account exists, recovery instructions have been sent",
		"recovery_token": token,
	})
}

func (d *DashboardHandler) handleRecoveryToken(w http.ResponseWriter, r *http.Request, token string) {
	if token == "" {
		writeAPIError(w, 400, "INVALID_TOKEN", "recovery token is required")
		return
	}

	rec, op, err := d.Auth.ValidateRecoveryToken(token)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			writeAPIError(w, 404, "INVALID_TOKEN", "recovery token not found or expired")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"valid":       true,
		"operator_id": op.ID,
		"tenant_id":   op.TenantID,
		"expires_at":  rec.ExpiresAt,
	})
}

func (d *DashboardHandler) handleGetAccountTOTP(w http.ResponseWriter, r *http.Request, tenantID, opID uuid.UUID) {
	op, err := d.Auth.GetOperatorByID(tenantID, opID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load operator")
		return
	}

	count, err := d.Auth.CountRemainingBackupCodes(opID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to count backup codes")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"totp_enabled":           op.TotpVerifiedAt != nil,
		"totp_verified_at":       op.TotpVerifiedAt,
		"totp_setup_required":    op.TotpSetupRequired,
		"backup_codes_remaining": count,
	})
}

func (d *DashboardHandler) handleRegenerateBackupCodes(w http.ResponseWriter, r *http.Request, opID uuid.UUID) {
	codes, err := d.Auth.RegenerateBackupCodes(opID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to regenerate backup codes")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"backup_codes": codes,
	})
}

func (d *DashboardHandler) handleListOperators(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	ops, err := d.Auth.ListOperators(tenantID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to list operators")
		return
	}
	if ops == nil {
		ops = []domain.Operator{}
	}
	WriteJSON(w, http.StatusOK, ops)
}

func (d *DashboardHandler) handleAdminResetTOTP(w http.ResponseWriter, r *http.Request, tenantID, adminID, targetOpID uuid.UUID) {
	setupToken, err := d.Auth.ResetOperatorTOTPByAdmin(tenantID, adminID, targetOpID)
	if err != nil {
		if errors.Is(err, storage.ErrUnauthorizedAdmin) {
			writeAPIError(w, 403, "FORBIDDEN", "admin privileges required")
			return
		}
		if errors.Is(err, storage.ErrOperatorNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "operator not found")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":      "reset",
		"operator_id": targetOpID,
		"setup_token": setupToken,
	})
}

func (d *DashboardHandler) handleAdminGetTOTPStatus(w http.ResponseWriter, r *http.Request, tenantID, adminID, targetOpID uuid.UUID) {
	admin, err := d.Auth.GetOperatorByID(tenantID, adminID)
	if err != nil || admin.Role != "admin" {
		writeAPIError(w, 403, "FORBIDDEN", "admin privileges required")
		return
	}

	op, err := d.Auth.GetOperatorByID(tenantID, targetOpID)
	if err != nil {
		if errors.Is(err, storage.ErrOperatorNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "operator not found")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to get operator")
		return
	}

	count, err := d.Auth.CountRemainingBackupCodes(targetOpID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to count backup codes")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"operator_id":            targetOpID,
		"totp_verified":          op.TotpVerifiedAt != nil,
		"totp_verified_at":       op.TotpVerifiedAt,
		"totp_setup_required":    op.TotpSetupRequired,
		"backup_codes_remaining": count,
	})
}

type createWhatsAppInvitationRequest struct {
	WhatsappNumber string `json:"whatsapp_number"`
	Whatsapp       string `json:"whatsapp"`
	Role           string `json:"role"`
}

func (d *DashboardHandler) handleCreateWhatsAppInvitation(w http.ResponseWriter, r *http.Request, tenantID, callerID uuid.UUID) {
	var req createWhatsAppInvitationRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	number := strings.TrimSpace(req.WhatsappNumber)
	if number == "" {
		number = strings.TrimSpace(req.Whatsapp)
	}
	if number == "" {
		writeAPIError(w, 400, "VALIDATION_ERROR", "whatsapp_number is required")
		return
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "operator"
	}

	inv, token, err := d.Auth.CreateInvitation(tenantID, &callerID, number, "whatsapp", role, number, "")
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	tenant, _ := d.Auth.GetTenantByID(tenantID)
	caller, _ := d.Auth.GetOperatorByID(tenantID, callerID)
	inviterName := caller.Name

	msgText := whatsapp.BuildInvitationMessage(tenant, inv, token, inviterName)

	var whatsappSent bool
	var sendErr error
	if d.WhatsApp != nil {
		if err := d.WhatsApp.SendInvitation(number, msgText); err != nil {
			sendErr = err
			_ = d.Auth.TrackInvitationDelivery(inv.ID, "failed", "", err.Error())
		} else {
			whatsappSent = true
			_ = d.Auth.TrackInvitationDelivery(inv.ID, "sent", "", "")
		}
	}

	resp := map[string]any{
		"invitation":    inv,
		"invite_token":  token,
		"invite_url":    "/dashboard/invitation/" + token,
		"whatsapp_sent": whatsappSent,
	}
	if !whatsappSent {
		if sendErr != nil {
			resp["whatsapp_error"] = sendErr.Error()
		} else {
			resp["whatsapp_error"] = "WhatsApp sender not configured or unavailable"
		}
		resp["manual_instructions"] = "Share this code manually: " + token
	}

	WriteJSON(w, http.StatusCreated, resp)
}

type createEmailInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (d *DashboardHandler) handleCreateEmailInvitation(w http.ResponseWriter, r *http.Request, tenantID, callerID uuid.UUID) {
	var req createEmailInvitationRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		writeAPIError(w, 400, "VALIDATION_ERROR", "email is required")
		return
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "operator"
	}

	inv, token, err := d.Auth.CreateInvitation(tenantID, &callerID, email, "email", role, "", email)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"invitation":   inv,
		"invite_token": token,
		"invite_url":   "/dashboard/invitation/" + token,
	})
}

func (d *DashboardHandler) handleListInvitations(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	invs, err := d.Auth.ListInvitations(tenantID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to list invitations")
		return
	}
	if invs == nil {
		invs = []domain.Invitation{}
	}
	WriteJSON(w, http.StatusOK, invs)
}

func (d *DashboardHandler) handleRevokeInvitation(w http.ResponseWriter, r *http.Request, tenantID, invID uuid.UUID) {
	err := d.Auth.RevokeInvitation(tenantID, invID)
	if err != nil {
		if errors.Is(err, storage.ErrInvitationNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "invitation not found")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"status": "revoked",
	})
}

func (d *DashboardHandler) handleGetTenantSetupStatus(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	status, err := d.Auth.GetTenantSetupStatus(tenantID)
	if err != nil {
		if errors.Is(err, storage.ErrTenantNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "tenant not found")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to get tenant setup status")
		return
	}
	WriteJSON(w, http.StatusOK, status)
}

type updateTenantSetupRequest struct {
	SetupStep  int            `json:"setup_step"`
	OrgDetails map[string]any `json:"org_details"`
}

func (d *DashboardHandler) handleUpdateTenantSetup(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	var req updateTenantSetupRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	status, err := d.Auth.UpdateTenantSetup(tenantID, req.SetupStep, req.OrgDetails)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, status)
}

func (d *DashboardHandler) handleCompleteTenantSetup(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	err := d.Auth.CompleteTenantSetup(tenantID)
	if err != nil {
		if errors.Is(err, storage.ErrTenantNotFound) {
			writeAPIError(w, 404, "NOT_FOUND", "tenant not found")
			return
		}
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"status":            "completed",
		"is_setup_complete": true,
	})
}
