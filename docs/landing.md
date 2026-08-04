# Landing

This document describes the marketing website. The website is in `apps/landing/`. It is an Astro static site.

## Overview

The landing page is the public marketing website for Vigil. It describes the product and gives install instructions. It is not part of the installed system. It has no connection to the backend.

The landing page is built with Astro 7. It uses Tailwind CSS 4. It uses Lucide for icons.

## Build and configuration

The Astro configuration is in `apps/landing/astro.config.mjs`. It configures the Vite plugin for Tailwind. It forces the PostCSS transformer for Tailwind v4.

The site has no integrations and no adapter. It is fully static.

The scripts are in `apps/landing/package.json`:

- `astro dev`: start the dev server.
- `astro build`: build the site.
- `astro preview`: preview the build.

## The routes

The site has three routes in `apps/landing/src/pages/`:

- `/` (`index.astro`): the homepage.
- `/privacy` (`privacy.astro`): the privacy page.
- `/docs` (`docs/index.astro`): a development placeholder.

There is no Starlight integration. The docs route is a placeholder that says "Still in development."

## The homepage

The homepage is the main marketing page. It is built from components in `apps/landing/src/components/`.

The headline is "Leave Vendor Lockin." The copy says that Vigil records RTSP/ONVIF cameras to your own hardware.

The hero has six capability tags:

- Watch on any screen you own.
- Runs on your hardware.
- No recurring charges.
- Open source.
- Your recordings stay on your hardware.
- Use Google Drive for backups.

The hero shows a DVR preview. It is a static illustration with four camera channels and a coverage timeline. It does not fetch camera data.

The install section gives the install command:

```bash
git clone https://github.com/jR4dh3y/vigil.git && cd vigil/deploy && docker compose up -d --build
```

The FAQ answers four questions about camera support, storage, remote access, and cost.

## The components

The main components are:

- `Header.astro`: the site header.
- `Hero.astro`: the hero section.
- `HeroPoints.astro`: the capability tags.
- `HangingOwl.astro`: the hanging owl art.
- `OwlMark.astro`: the Vigil owl mark.
- `DvrPreview.astro`: the DVR preview illustration.
- `SystemParts.astro`: the install and system section.
- `Faq.astro`: the FAQ.
- `Footer.astro`: the footer.

Some components are present but not referenced by the active page tree. These are `ProofStrip.astro`, `Pipeline.astro`, and `Storage.astro`. They are source material for future sections.

## The owl art

The site uses owl images as the brand mark. The images are in `apps/landing/public/assets/`. There are many owl poses, for example:

- `owl.png`: the default mark.
- `owl-pointer-right.png`: the hero pointing owl.
- `owl-moon-branch.png`: the install owl.
- `owl-question.png`: the FAQ owl.
- `owl-construction.png`: the docs placeholder.

The hero uses a composited owl over a night scene. It is not a pure image generation.

## Styling

The site uses a paper-and-ink theme. The background is cream (`#f3eee3`). The ink is dark (`#1e1e20`). The accent is lavender (`#c59edc`).

The install section is dark. The DVR preview uses the CCTV-monitor motif with dark feed tiles and status indicators.

The body background has a 24-pixel grid pattern. The site uses Arial and monospace font stacks. There is no IBM Plex import.

## Client-side behavior

The client script is in `apps/landing/src/scripts/main.ts`. It:

- Updates the local clock every second.
- Copies the install command on button press.
- Coordinates the hero scroll animation.

The hero choreography uses `requestAnimationFrame`. It respects the `prefers-reduced-motion` setting.

The site has no fetch calls and no API client. It reads no environment variables.

## The privacy page

The privacy page states that camera streams and recordings stay on the user's server. It says the connection secret is encrypted at rest. It says the static website sets no cookies and runs no analytics.

The page provides the contact email and the GitHub issues page.

## The footer

The footer shows the Vigil brand, a GitHub source link, a privacy link, and the author credit.