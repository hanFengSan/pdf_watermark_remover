中文版：[Link](https://github.com/hanFengSan/pdf_watermark_remover/blob/main/README_CN.md)
# PDF Watermark Remover

A local tool that renders PDF pages into images, performs watermark suppression and readability restoration, and finally reconstructs a black-and-white output PDF. The current processing pipeline is specifically optimized for text-based PDFs.

## Usage Guide

1. Place the executable file (`.exe`) and the PDF(s) to be processed in the same directory.
   - **Note:** Files containing `remove_watermark` in their filenames will be automatically skipped.
2. Run the program (processing 100 pages takes approximately 3 minutes, depending on hardware configuration and document complexity).
3. Check the current directory for the output results.
4. **Other Usage:**
   - Specify a single PDF via command line: `pdf_watermark_remover xx/path/xxx.pdf`
   - The output file will be generated in the same directory as the source PDF.

## Processing Pipeline

1. Validate input PDF.
2. Render each page as a PNG image (default 400 DPI).
3. Estimate watermark patterns and select suppression strategies.
4. Execute watermark suppression, binarization, and readability cleanup.
5. Apply retry strategies for pages with significant ghosting/artifacts.
6. Reconstruct the processed images into the output PDF.

## Build

**Prerequisites:**

- Go 1.25+

**Build for current platform:**

```bash
go build -o pdf_watermark_remover ./cmd/pdf_watermark_remover
```

**Build for Windows amd64:**

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/pdf_watermark_remover_windows_amd64.exe ./cmd/pdf_watermark_remover
```