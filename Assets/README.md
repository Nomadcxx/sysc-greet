# Assets Directory

This directory contains bundled video assets for background effects.

## Video Files

Place the following video files in this directory:

- **Fireplace.mp4** - Cozy fireplace animation for Fireplace background effect
- **ParticleEffects.mp4** - Particle field animation for Particle Field background effect

## Usage

These assets are automatically detected and used by the greeter:
- When "Fireplace" background is selected, it plays `Assets/Fireplace.mp4` via gslapper
- When "Particle Field" background is selected, it plays `Assets/ParticleEffects.mp4` via gslapper (if available)

## Installation

The installer will copy this Assets directory to `/usr/share/sysc-greet/Assets/` for production use.
