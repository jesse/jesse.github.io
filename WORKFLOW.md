# Adding photos and publishing

This is for the photographer. You do not need to know how the website is built.
Adding a photo is: export a JPG, put it in the `photos` folder, save to GitHub. A
couple of minutes later it is live.

## One-time setup

You need a free GitHub account with access to the repository. Optionally, GitHub Desktop
(free) makes the "put it in the folder and save" step a drag and drop on your computer.

The site is hosted by GitHub Pages, also free. Every time something changes on the
`main` branch, GitHub rebuilds the site and publishes it. You never upload anywhere else.

## Exporting photos

Export from Lightroom, Capture One or Photos with these settings:

| Setting | Value |
| --- | --- |
| Format | JPEG |
| Color space | sRGB |
| Resize to fit | Long edge, 3000 pixels |
| Quality | 80 (Lightroom) or about 85% elsewhere |
| Sharpening | For screen, standard |
| Metadata | Anything. The build strips camera data and location and stamps your copyright. |

A file exported this way is 1 to 3 MB. Bigger files work, but they make the repository
heavier for no visible gain, and the build prints a note when a file is over 6 MB.

Name the file anything you like. The name is not shown on the site.

Portrait and landscape both work. Phone photos that rely on a rotation flag come out
upright. Every JPG in `photos` is in the slideshow, in random order; there are no
categories.

## Adding photos

### In the browser (no software needed)

1. Open the repository on github.com and click the `photos` folder.
2. Click "Add file", then "Upload files".
3. Drag your JPGs onto the page.
4. At the bottom, keep "Commit directly to the main branch" selected and click
   "Commit changes".

GitHub's uploader accepts up to 100 files per batch and 25 MB per file.

### With GitHub Desktop

1. Copy the JPGs into the `photos` folder on your computer.
2. In GitHub Desktop, type a short summary like "Add June landscapes" and click
   "Commit to main".
3. Click "Push origin".

## What happens next

GitHub notices the change and runs the build. The build resizes every new photo to four
sizes, writes the list of photos into the page, and publishes. Expect 1 to 3 minutes.

You can watch progress on github.com under the repository's "Actions" tab. A green
check means it is live. Reload the site; the new photos are in the random rotation.
GitHub caches pages for ten minutes, so a reload right after the green check may
briefly show the old set.

## Removing or replacing a photo

- **Remove**: delete the file from `photos` (on github.com, open the file and use the
  trash icon; or delete it on disk and commit in GitHub Desktop). The next build drops it.
- **Replace**: overwrite the file with the new version under the same name and commit.
  The site treats it as a new photo and every visitor sees the new one; no cache issues.

## Changing your name or booking link

Edit `site.config.json` in the repository (on github.com, open it and click the pencil):

```json
{
  "name": "Jesse Young",
  "title": "Jesse Young Photography",
  "bookingLabel": "Book a session",
  "bookingUrl": "mailto:youngj@gmail.com?subject=Jesse%20Young%20Photography%20inquiry",
  "intervalMs": 5000,
  "fadeMs": 1000
}
```

`title` is shown top-left on the page and in the browser tab. `name` is the copyright
line, and goes into the copyright field inside every image file the site serves.
`bookingLabel` and `bookingUrl` are the white button bottom-right; the URL can
be a `mailto:` address or a link to a booking page. `intervalMs` is how long each photo stays,
in milliseconds (5000 is five seconds). `fadeMs` is how long the crossfade between two photos takes (1000 is one second). Commit the change and the site rebuilds.

## Previewing on your computer (optional)

Only if you want to see a change before publishing. Needs Node.js (free, nodejs.org).

```
npm install        # first time only
npm run build      # resizes photos into dist/
npm run preview    # then open http://localhost:8000
```

## If something goes wrong

- **Build failed** (red X under Actions): open the run and read the last lines of the
  log. The usual causes are a file that is not actually a JPG (a `.png` or `.heic`
  renamed), or a file over 25 MB. Fix the file and commit again.
- **A photo looks wrong (sideways, wrong colors)**: re-export it with the settings above
  and replace it.
- **An old photo still shows after replacing it**: wait ten minutes for GitHub's cache,
  then hard reload (Shift + reload). Image file names change whenever the file changes.
- **The site is slow to show the first photo**: the first photo is the largest download
  on the page. Check that exports are not much larger than 3000 px long edge.

## Privacy and scraping, briefly

The site sets no cookies and loads nothing from any other company. It asks AI crawlers
to stay away through `robots.txt`, `ai.txt` and a tag in the page. These are requests,
not walls; GitHub Pages has no setting to block bots. Nothing you add in `photos`
changes any of this.
