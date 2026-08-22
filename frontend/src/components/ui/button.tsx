import { ButtonHTMLAttributes, forwardRef } from 'react'
import { twMerge } from 'tailwind-merge'

export type ButtonVariant = 'default' | 'primary' | 'secondary' | 'ghost' | 'danger'
export type ButtonSize = 'sm' | 'md' | 'lg' | 'xl'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'md', ...props }, ref) => {
    const baseClasses = 'inline-flex items-center justify-center gap-1.5 rounded-xl font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-orange-500/40 disabled:opacity-50 disabled:pointer-events-none'

    const variantClasses = {
      default: 'bg-white text-gray-900 border border-gray-200 shadow-sm hover:bg-[#f8f4ef] hover:border-gray-300',
      primary: 'bg-orange-500 text-white shadow-sm hover:bg-orange-600',
      secondary: 'bg-white/80 text-gray-800 border border-gray-300 hover:bg-white',
      ghost: 'bg-transparent text-gray-700 hover:bg-white/80 hover:text-gray-950',
      danger: 'bg-red-600 text-white shadow-sm hover:bg-red-700',
    }

    const sizeClasses = {
      sm: 'px-2.5 py-1 text-xs',
      md: 'px-3 py-1.5 text-sm',
      lg: 'px-4 py-2 text-sm',
      xl: 'px-8 py-3.5 text-base rounded-xl',
    }

    return (
      <button
        className={twMerge(
          baseClasses,
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)

Button.displayName = 'Button'

export default Button