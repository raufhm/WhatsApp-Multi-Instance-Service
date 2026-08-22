import { HTMLAttributes, forwardRef } from 'react'
import { twMerge } from 'tailwind-merge'

export type CardVariant = 'default' | 'elevated'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const baseClasses = 'rounded-2xl border border-white/75 bg-white/80 backdrop-blur-xl'

    const variantClasses = {
      default: 'shadow-[0_18px_50px_-32px_rgba(15,23,42,0.55)]',
      elevated: 'shadow-[0_24px_60px_-30px_rgba(15,23,42,0.6)]',
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