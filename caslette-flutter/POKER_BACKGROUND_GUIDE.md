# Poker Table Background Optimization Guide

## Current Issues Fixed

1. **Poor scaling across devices** - The original single texture with `BoxFit.cover` and `scale: 0.5` doesn't adapt well
2. **Large file size** - Your current texture is ~4.7MB, which is too large for mobile
3. **No device density support** - Missing 2x and 3x variants for high-DPI screens
4. **Fixed scaling** - No responsive behavior for different screen sizes and orientations

## Recommended Asset Structure

Create the following directory structure in `assets/images/`:

```
assets/images/
├── poker_table_bg.png          # Base 1x density (1024x768px max)
├── felt_pattern.png            # Small tileable pattern (64x64px)
├── 2.0x/
│   └── poker_table_bg.png      # 2x density (2048x1536px max)
└── 3.0x/
    └── poker_table_bg.png      # 3x density (3072x2304px max)
```

## Asset Specifications

### Main Background (`poker_table_bg.png`)

- **Format**: PNG with transparency or JPEG for solid backgrounds
- **1x version**: 1024x768px (4:3 aspect ratio) - ~200-500KB
- **2x version**: 2048x1536px - ~500KB-1MB
- **3x version**: 3072x2304px - ~800KB-1.5MB
- **Colors**: Deep green felt colors (#0F5132, #1F7A44)
- **Design**: Oval poker table shape centered, with subtle felt texture

### Felt Pattern (`felt_pattern.png`)

- **Format**: PNG with transparency
- **Size**: 64x64px or 128x128px
- **File size**: <50KB
- **Purpose**: Tileable texture overlay for programmatic backgrounds

## Image Optimization Tips

### 1. Compression

- Use tools like TinyPNG, ImageOptim, or Squoosh
- Target 70-80% quality for JPEG, optimize PNG with palette reduction
- Aim for 1x version under 500KB, 2x under 1MB

### 2. Design Considerations

- Keep important visual elements (table edge, center area) in the middle 80% of the image
- Use gradients from center to edges to accommodate different crop areas
- Avoid fine details that won't be visible on small screens
- Ensure good contrast with white UI elements

### 3. Creating Responsive Designs

- Design with 16:9, 4:3, and 3:2 aspect ratios in mind
- Make the center area (where cards/pot display) visually neutral
- Use radial gradients that work well when cropped to different shapes

## Implementation Options

### Option 1: Image-Based (Recommended for Rich Visuals)

```dart
PokerTableBackground(
  useGradientFallback: false,
  child: yourGameContent,
)
```

### Option 2: Programmatic (Recommended for Performance)

```dart
ProgrammaticPokerBackground(
  primaryColor: Color(0xFF0F5132),
  accentColor: Color(0xFF1F7A44),
  child: yourGameContent,
)
```

### Option 3: Hybrid Approach

Use programmatic background with small pattern overlay:

```dart
PokerTableBackground(
  useGradientFallback: true, // Uses gradient + felt pattern
  child: yourGameContent,
)
```

## Testing Different Devices

Test your background on:

- Phone portrait (9:16, 10:16)
- Phone landscape (16:9, 16:10)
- Tablet portrait (4:3, 3:4)
- Tablet landscape (4:3, 16:10)

## Migration Steps

1. **Create optimized assets** following the specifications above
2. **Test with image-based approach** using `PokerTableBackground`
3. **Fallback option**: If images cause issues, switch to `ProgrammaticPokerBackground`
4. **Remove old asset**: Delete `poker_table_texture.png` once satisfied

## Asset Creation Tools

- **Figma/Sketch**: For designing the table layout
- **GIMP/Photoshop**: For texture creation and export
- **Squoosh.app**: For web-based compression
- **ImageOptim**: For batch optimization on Mac
- **TinyPNG**: For online PNG compression

## Performance Considerations

- **Memory usage**: Multiple density variants use more memory but provide crisp visuals
- **Loading time**: Smaller files load faster, especially on slow connections
- **Battery impact**: Programmatic backgrounds use less memory but more CPU
- **Caching**: Flutter automatically caches image assets

Choose the approach that best fits your app's performance requirements and visual goals.
