# Kitty Mesh Background Design

## Goal

Preserve per-dot model shading in the controlled Kitty greeter while keeping sysc-Go's existing ANSI renderer available to all other consumers.

## Approaches considered

1. **Kitty image layer (chosen).** Export the existing mesh renderer's transparent per-dot image and let sysc-greet present it through Kitty's graphics protocol. This preserves the current 2×4 dot resolution and confines terminal-specific behavior to the greeter.
2. **Change shared cell composition.** Two-color Braille and half-block output retain some color variation, but the fixed-frame comparisons show patches or softened thin features. A shared default would also alter LogoSpin and other sysc-Go callers.
3. **Replace or extend the renderer with triangle/GPU rendering.** Higher-resolution references show more detail, but the current evidence does not identify source sampling or compute as the first limitation. This adds much more renderer and packaging scope.

## Design

sysc-Go will add an additive mesh-frame API that returns an `image.NRGBA` at two columns by four rows of pixels per terminal cell, with alpha zero for unoccupied dots, plus the existing attribution line. The mesh projection, depth test, lighting, frame timing and model data remain shared with `Render() string`; ANSI output remains unchanged. The renderer will not emit terminal sequences.

sysc-greet will use the image API only when Kitty sets `KITTY_WINDOW_ID` and one of the model backgrounds is active. It will encode the frame with Go's PNG encoder and send chunked Kitty graphics commands through Bubble Tea's `tea.Raw` output path. The image is placed behind text and non-default cell backgrounds. During this path the TUI uses Kitty's configured `#1a1a1a` default background, matching sysc-greet's existing `BgBase`; other terminal runs keep the current ANSI view and background color.

The greeter owns the image IDs, placement size, updates, deletion on background/mode changes, and cleanup at exit. It refreshes the image after mesh ticks and terminal resizes. The existing credit remains ordinary TUI text above the image. Pixel cell dimensions continue to feed `SetCellAspect` at model creation and resize. No terminal protocol code enters `Render()`, no new dependency is added, and no sample-density or GPU change is included.

## Tradeoffs and limits

- Kitty gets independent color for each existing Braille dot; other sysc-Go consumers retain ANSI output until they choose their own presentation path.
- PNG encoding and terminal writes add work per frame. Verify output cost at the greeter's real viewport and retain the existing 491×90 renderer budget.
- This does not add triangle coverage, material/texture preservation, or detail beyond the current 2×4 dot grid.
- If Kitty is not detected, sysc-greet continues to render the current ANSI frame.

## Verification

- Verify one image cell preserves distinct occupied-dot colors and leaves unoccupied dots transparent.
- Verify Kitty command chunking, placement, replacement, and deletion with focused unit checks.
- Run the sysc-Go animation and sysc-greet package checks, plus the existing 491×90 budget check.
- Capture the greeter in Kitty at 40×12, 60×18 and 80×24, checking the model, login form, credit line, resize, and switching away from the background. Do not restart the live greetd session; use an isolated Kitty session and leave the live check for the user's next reboot.
