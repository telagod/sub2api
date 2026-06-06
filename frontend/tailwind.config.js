import animate from 'tailwindcss-animate'

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // ── shadcn 语义层（冷钢黑白，HSL CSS 变量驱动，见 style.css :root/.dark）──
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))'
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))'
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))'
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))'
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))'
        },

        // ── primary：铬银强调（替换旧 teal）。色阶单调，300=高光银，500=中钢 ──
        primary: {
          50: '#f6f7f8',
          100: '#eceef1',
          200: '#dadde2',
          300: '#c0c4cc', // 铬银高光
          400: '#a4a9b3',
          500: '#888e99', // 中钢（主强调）
          600: '#6c727d',
          700: '#565a63',
          800: '#3f4248',
          900: '#2a2c30',
          950: '#1a1b1e',
          foreground: 'hsl(var(--primary-foreground))',
          DEFAULT: '#c0c4cc'
        },

        // ── accent：金属银（hover / 次强调）──
        accent: {
          50: '#f8fafc',
          100: '#eceef1',
          200: '#dadde2',
          300: '#bcc1c9',
          400: '#9aa0a9',
          500: '#7c828c',
          600: '#5e636c',
          700: '#474b52',
          800: '#33363b',
          900: '#222427',
          950: '#141517',
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))'
        },

        // ── dark：纯炭中性阶（去 slate 蓝），用于 surface/border/bg ──
        dark: {
          50: '#f5f6f7',
          100: '#e8eaeb',
          200: '#c9cccf',
          300: '#9da2a7',
          400: '#6e7378',
          500: '#4a4e53',
          600: '#2f3236',
          700: '#26282b', // border
          800: '#1a1c1e', // surface
          900: '#121315', // elevated bg
          950: '#0a0a0b' // page bg
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      borderRadius: {
        // 收紧圆角，硬朗金属感；lg/md/sm 绑定 --radius 供 shadcn 组件
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
        '4xl': '2rem'
      },
      boxShadow: {
        // ── 金属质感阴影：冷投影 + 顶部高光边（inset）──
        metal: '0 1px 2px rgba(0,0,0,.6), 0 4px 16px rgba(0,0,0,.45)',
        'metal-sm': '0 1px 2px rgba(0,0,0,.5)',
        'metal-lg': '0 2px 4px rgba(0,0,0,.6), 0 12px 40px rgba(0,0,0,.5)',
        'metal-edge': 'inset 0 1px 0 rgba(255,255,255,.10), inset 0 -1px 0 rgba(0,0,0,.4)',
        'metal-edge-strong': 'inset 0 1px 0 rgba(255,255,255,.18), inset 0 -1px 0 rgba(0,0,0,.5)',
        'metal-ring': '0 0 0 1px rgba(255,255,255,.06)',
        // 兼容旧名（玻璃/光晕 → 退化为冷阴影，避免残留页面报错）
        glass: '0 8px 32px rgba(0,0,0,.45)',
        'glass-sm': '0 4px 16px rgba(0,0,0,.35)',
        glow: '0 0 0 1px rgba(255,255,255,.08)',
        'glow-lg': '0 0 0 1px rgba(255,255,255,.12)',
        card: '0 1px 2px rgba(0,0,0,.5), 0 1px 3px rgba(0,0,0,.4)',
        'card-hover': '0 2px 4px rgba(0,0,0,.6), 0 12px 32px rgba(0,0,0,.5)',
        'inner-glow': 'inset 0 1px 0 rgba(255,255,255,.10)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        // 金属表面渐变（gunmetal）
        'metal-surface': 'linear-gradient(180deg, #1f2225 0%, #16181a 100%)',
        'metal-raised': 'linear-gradient(180deg, #2a2d31 0%, #1c1f22 100%)',
        // 亮银主按钮（实体金属）
        'metal-silver': 'linear-gradient(180deg, #e9ebee 0%, #c0c4cc 48%, #a9aeb6 100%)',
        'metal-silver-hover': 'linear-gradient(180deg, #f2f4f6 0%, #ccd0d6 48%, #b6bbc3 100%)',
        // 拉丝高光扫光
        'metal-sheen':
          'linear-gradient(105deg, transparent 40%, rgba(255,255,255,.08) 50%, transparent 60%)',
        // 兼容旧名
        'gradient-primary': 'linear-gradient(180deg, #e9ebee 0%, #c0c4cc 48%, #a9aeb6 100%)',
        'gradient-dark': 'linear-gradient(135deg, #1c1f22 0%, #0a0a0b 100%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out'
      },
      keyframes: {
        fadeIn: { '0%': { opacity: '0' }, '100%': { opacity: '1' } },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        'accordion-down': {
          from: { height: '0' },
          to: { height: 'var(--reka-accordion-content-height)' }
        },
        'accordion-up': {
          from: { height: 'var(--reka-accordion-content-height)' },
          to: { height: '0' }
        }
      },
      backdropBlur: { xs: '2px' }
    }
  },
  plugins: [animate]
}
