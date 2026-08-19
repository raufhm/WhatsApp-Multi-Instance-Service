import { LabelHTMLAttributes, forwardRef } from 'react'
import { twMerge } from 'tailwind-merge'

interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {}

export const Label = forwardRef<HTMLLabelElement, LabelProps>(
  ({ className, ...props }, ref) => {
    return (
      <label
        ref={ref}
        className={twMerge('block text-[13px] font-medium text-gray-600 mb-1', className)}
        {...props}
      />
    )
  }
)

Label.displayName = 'Label'

export default Label