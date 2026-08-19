import { InputHTMLAttributes, forwardRef } from 'react'
import { twMerge } from 'tailwind-merge'

export type InputSize = 'sm' | 'md' | 'lg'

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  size?: InputSize
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, size = 'md', ...props }, ref) => {
    const baseClasses = 'px-3 py-1.5 border border-gray-300 rounded-md bg-white shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 disabled:opacity-50'

    const sizeClasses = {
      sm: 'text-sm px-2 py-1',
      md: 'text-sm px-3 py-1.5',
      lg: 'text-base px-4 py-2',
    }

    return (
      <input
        className={twMerge(baseClasses, sizeClasses[size], className)}
        ref={ref}
        {...props}
      />
    )
  }
)

Input.displayName = 'Input'

export default Input