import { cva, type VariantProps } from 'class-variance-authority'

export { default as Button } from './Button.vue'

/**
 * 冷钢黑白金属按钮变体（shadcn-vue 风格 + COLD STEEL 皮肤）。
 * default = 亮银实体金属（深字 + 顶部高光边），其余为 gunmetal / 描边 / 语义。
 */
export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/60 focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 active:scale-[0.985]',
  {
    variants: {
      variant: {
        default:
          'bg-metal-silver text-dark-950 font-semibold border border-white/25 [box-shadow:inset_0_1px_0_rgba(255,255,255,.55),0_1px_2px_rgba(0,0,0,.55)] hover:bg-metal-silver-hover',
        secondary:
          'bg-metal-raised text-foreground border border-border shadow-metal-edge hover:brightness-110',
        outline:
          'border border-border bg-transparent text-foreground hover:bg-accent hover:text-foreground',
        ghost: 'text-muted-foreground hover:bg-accent hover:text-foreground',
        destructive:
          'text-white [background-image:linear-gradient(180deg,#c2453f_0%,#9b2f2a_100%)] [box-shadow:inset_0_1px_0_rgba(255,255,255,.16),0_1px_2px_rgba(0,0,0,.5)] hover:brightness-110',
        link: 'text-primary-200 underline-offset-4 hover:underline'
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 rounded px-3 text-xs',
        lg: 'h-11 rounded-lg px-6 text-base',
        icon: 'h-10 w-10'
      }
    },
    defaultVariants: { variant: 'default', size: 'default' }
  }
)

export type ButtonVariants = VariantProps<typeof buttonVariants>
