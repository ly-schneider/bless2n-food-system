import { cva, type VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn } from "@/lib/utils"

const iconButtonVariants = cva(
  "inline-flex items-center justify-center transition-all disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        outline: "border border-border bg-card text-card-foreground hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
      },
      size: {
        sm: "h-8 w-8",
        default: "h-10 w-10",
        lg: "h-12 w-12",
        xl: "h-16 w-16",
      },
      shape: {
        square: "rounded-md",
        circle: "rounded-full",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
      shape: "square",
    },
  }
)

interface IconButtonProps extends React.ComponentPropsWithoutRef<"button">, VariantProps<typeof iconButtonVariants> {
  asChild?: boolean
}

interface IconLinkProps extends React.ComponentPropsWithoutRef<"a">, VariantProps<typeof iconButtonVariants> {}

const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ className, variant, size, shape, ...props }, ref) => {
    return <button className={cn(iconButtonVariants({ variant, size, shape, className }))} ref={ref} {...props} />
  }
)
IconButton.displayName = "IconButton"

const IconLink = React.forwardRef<HTMLAnchorElement, IconLinkProps>(
  ({ className, variant, size, shape, ...props }, ref) => {
    return <a className={cn(iconButtonVariants({ variant, size, shape, className }))} ref={ref} {...props} />
  }
)
IconLink.displayName = "IconLink"

export { IconButton, IconLink, iconButtonVariants }
