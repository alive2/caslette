// Cosmic theme constants matching the game's aesthetic
export const cosmicTheme = {
  colors: {
    // Background gradients
    background: {
      primary: "from-slate-900 via-purple-900 to-slate-900",
      secondary: "from-gray-900 via-blue-900 to-gray-900",
      card: "from-slate-800/50 to-purple-900/30",
      hover: "from-slate-700/60 to-purple-800/40",
    },

    // Text colors
    text: {
      primary: "text-white",
      secondary: "text-gray-300",
      accent: "text-purple-400",
      muted: "text-gray-500",
    },

    // Accent colors
    accent: {
      primary: "text-purple-400 bg-purple-500/20",
      secondary: "text-blue-400 bg-blue-500/20",
      success: "text-green-400 bg-green-500/20",
      warning: "text-yellow-400 bg-yellow-500/20",
      danger: "text-red-400 bg-red-500/20",
    },

    // Button styles
    button: {
      primary:
        "bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white",
      secondary:
        "bg-gradient-to-r from-gray-700 to-gray-800 hover:from-gray-600 hover:to-gray-700 text-white",
      success:
        "bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700 text-white",
      danger:
        "bg-gradient-to-r from-red-600 to-pink-600 hover:from-red-700 hover:to-pink-700 text-white",
      ghost: "bg-white/10 hover:bg-white/20 text-white border border-white/20",
    },

    // Border and divider colors
    border: {
      primary: "border-white/20",
      secondary: "border-gray-700",
      accent: "border-purple-500/50",
    },

    // Input styles
    input: {
      base: "bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-purple-500/50 focus:ring-purple-500/20",
      error: "bg-red-500/10 border-red-500/50 text-white",
    },

    // Card and panel styles
    panel: {
      base: "bg-gradient-to-br from-slate-800/80 to-purple-900/50 backdrop-blur-sm border border-white/10",
      elevated:
        "bg-gradient-to-br from-slate-700/90 to-purple-800/60 backdrop-blur-md border border-white/20 shadow-2xl",
    },

    // Status indicators
    status: {
      active: "bg-green-500/20 text-green-400 border-green-500/30",
      inactive: "bg-red-500/20 text-red-400 border-red-500/30",
      pending: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    },
  },

  // Animation and effects
  effects: {
    glow: "shadow-lg shadow-purple-500/25",
    glowHover: "hover:shadow-xl hover:shadow-purple-500/40",
    shimmer: "animate-pulse",
    transition: "transition-all duration-300",
  },

  // Layout constants
  layout: {
    sidebarWidth: "w-64",
    borderRadius: "rounded-lg",
    spacing: {
      xs: "p-2",
      sm: "p-4",
      md: "p-6",
      lg: "p-8",
    },
  },
};

// Helper function to combine cosmic classes
export const cosmic = {
  card: `${cosmicTheme.colors.panel.base} ${cosmicTheme.layout.borderRadius} ${cosmicTheme.layout.spacing.md}`,
  cardElevated: `${cosmicTheme.colors.panel.elevated} ${cosmicTheme.layout.borderRadius} ${cosmicTheme.layout.spacing.md} ${cosmicTheme.effects.glow}`,
  button: {
    primary: `${cosmicTheme.colors.button.primary} ${cosmicTheme.layout.borderRadius} px-4 py-2 font-medium ${cosmicTheme.effects.transition}`,
    secondary: `${cosmicTheme.colors.button.secondary} ${cosmicTheme.layout.borderRadius} px-4 py-2 font-medium ${cosmicTheme.effects.transition}`,
    success: `${cosmicTheme.colors.button.success} ${cosmicTheme.layout.borderRadius} px-4 py-2 font-medium ${cosmicTheme.effects.transition}`,
    danger: `${cosmicTheme.colors.button.danger} ${cosmicTheme.layout.borderRadius} px-4 py-2 font-medium ${cosmicTheme.effects.transition}`,
    ghost: `${cosmicTheme.colors.button.ghost} ${cosmicTheme.layout.borderRadius} px-4 py-2 font-medium ${cosmicTheme.effects.transition}`,
  },
  input: `${cosmicTheme.colors.input.base} ${cosmicTheme.layout.borderRadius} px-3 py-2 ${cosmicTheme.effects.transition}`,
  select: `${cosmicTheme.colors.input.base} ${cosmicTheme.layout.borderRadius} px-3 py-2 ${cosmicTheme.effects.transition} appearance-none bg-no-repeat bg-right bg-[length:12px] bg-[url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='m6 8 4 4 4-4'/%3e%3c/svg%3e")] pr-10`,
  background: `bg-gradient-to-br ${cosmicTheme.colors.background.primary}`,
};
