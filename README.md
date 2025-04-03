Performance:
Let me clarify the complete rendering picture and why this approach is still efficient in practice:

### Full Rendering Pipeline (per frame)

1. **Complete Redraw**:
   - The entire chart (axes, OHLC bars, grid, labels) is redrawn every frame (~16.67ms at 60FPS)
   - This includes:
     * Background fill
     * Grid segments (your new alternating dark/light areas)
     * Hundreds of OHLC bars
     * Dozens of price/time labels
     * Crosshair (when visible)

2. **Why This Works**:
   - **Ebitengine's Optimization**: The engine batches draw calls efficiently
   - **GPU Acceleration**: Modern GPUs handle 2D rendering of this scale easily
   - **Minimal State Changes**: Your chart uses consistent styling (same colors/fonts)

### Performance Breakdown

| Element          | Quantity | Relative Cost | Notes |
|------------------|----------|---------------|-------|
| OHLC Bars        | ~300     | Medium        | Simple vector lines |
| Grid Segments    | ~24      | Low           | Large filled rects |
| Price Labels     | ~10      | Medium        | Text rendering |
| Time Labels      | ~12      | Medium        | Text rendering |
| Crosshair        | 2 lines + 2 labels | Low | Simple elements |

### Optimization Reality Check

1. **Text Rendering** is typically the most expensive part, but:
   - You're using bitmap fonts (`basicfont.Face7x13`)
   - Limited quantity (dozens, not thousands)

2. **Vector Operations** (lines, rects) are cheap:
   ```go
   // Example: Drawing an OHLC bar costs ~0.01ms
   vector.StrokeLine(screen, x, top, x, bottom, width, color, false)
   ```

3. **Typical Performance**:
   - On modern hardware: 0.5-3ms per frame for this chart
   - Leaves 13-16ms for other game logic (plenty of headroom)

### When You Might Need Optimization

1. **If Rendering Slows Down**:
   - Implement dirty-rectangle rendering (only update changed areas)
   - Cache rendered chart to texture (only redraw when data changes)
   - Use level-of-detail (simpler rendering when zoomed out)

2. **Current Code Doesn't Need It Because**:
   ```go
   // In main.go's Draw():
   g.axes.Draw(screen, g.chart)    // Redraws everything
   g.chart.Draw(screen)            // Redraws all bars
   g.interaction.Draw(screen, g.chart) // Redraws crosshair
   ```
   - Even with full redraw, stays well within frame budget
   - Simpler code maintenance vs. complex partial updates

### Advanced Alternative (If Needed)

For extreme cases (10,000+ elements), you could:

1. **Cache Static Elements**:
   ```go
   // Create once when data changes
   chartTexture := ebiten.NewImage(width, height)
   vector.DrawFilledRect(chartTexture, ...) // Draw grid
   // etc...

   // Then in Draw():
   screen.DrawImage(chartTexture, nil) // Fast copy
   ```

2. **Dynamic Elements Separately**:
   ```go
   // Only redraw these each frame
   crosshairTexture := ebiten.NewImage(width, height)
   // Draw crosshair to texture
   screen.DrawImage(crosshairTexture, nil)
   ```

But your current implementation is correct for the scale shown - the "full redraw" approach is standard for financial charts at this complexity level. The crosshair addition is negligible to the overall rendering cost.