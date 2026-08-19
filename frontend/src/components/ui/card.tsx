import { HTMLAttributes, forwardRef } from 'react'
import { twMerge } from 'tailwind-merge'

export type CardVariant = 'default' | 'elevated'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const baseClasses = 'rounded-lg border bg-white'

    const variantClasses = {
      default: 'border-gray-200/80 shadow-sm',
      elevated: 'border-gray-200 shadow-md',
    }

    return (
      <div
        ref={ref}
        className={twMerge(baseClasses, variantClasses[variant], className)}
        {...props}
      />
    )
  }
)

Card.displayName = 'Card'

export default Card