UPDATE invitations SET role = LOWER(TRIM(role)) WHERE role IS NOT NULL;
UPDATE operators SET role = LOWER(TRIM(role)) WHERE role IS NOT NULL;
