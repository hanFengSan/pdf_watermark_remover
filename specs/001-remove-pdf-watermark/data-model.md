# Data Model: PDF Watermark Removal CLI

## Entity: InputDocument

- Description: Source PDF provided by user in CLI arguments.
- Fields:
  - `path` (string, required): Absolute or relative file path.
  - `exists` (bool): File presence check result.
  - `readable` (bool): Read permission and open success.
  - `is_pdf` (bool): Basic format validation result.
  - `page_count` (int): Number of pages; must be >=2.
- Validation rules:
  - Must exist, be readable, and pass PDF validation.
  - `page_count >= 2` required for processing.

## Entity: PageImage

- Description: Per-page rasterized image and processing state.
- Fields:
  - `page_index` (int, required): 1-based index in source order.
  - `source_image_path` (string, required): Path to rasterized page image.
  - `processed_image_path` (string, optional): Path to transformed output page image.
  - `orientation` (string): Normalized orientation metadata.
  - `watermark_confidence` (float): Confidence score of watermark match for this page.
  - `processing_mode` (enum): `standard` | `best_effort`.
  - `status` (enum): `pending` | `processed` | `failed`.
- Validation rules:
  - Every input page must produce one `PageImage` record.
  - `page_index` must remain unique and ordered.

## Entity: WatermarkPattern

- Description: Document-level inferred watermark pattern used to guide suppression.
- Fields:
  - `angle_estimate` (float): Dominant diagonal angle estimate.
  - `intensity_band` (string): Expected lightness band for watermark pixels.
  - `spatial_tendency` (string): Typical placement tendency across pages.
  - `pattern_confidence` (float): Confidence of shared pattern inference.
  - `variation_flag` (bool): Whether inter-page inconsistency is detected.
- Validation rules:
  - Must be produced before page suppression begins.
  - Low confidence sets page-level `processing_mode = best_effort`.

## Entity: OutputDocument

- Description: Final generated PDF and execution result metadata.
- Fields:
  - `output_path` (string, required): Final file path.
  - `naming_mode` (enum): `default_suffix` | `auto_incremented`.
  - `page_count` (int): Must equal input page count on success.
  - `created` (bool): Whether file was successfully written.
  - `duration_seconds` (float): End-to-end runtime.
  - `exit_code` (int): CLI process result.
- Validation rules:
  - Must not overwrite input file.
  - If default output path exists, must select next available incremented name.
  - On successful processing, `page_count` equals `InputDocument.page_count`.

## Relationships

- `InputDocument` 1 -> N `PageImage`
- `InputDocument` 1 -> 1 `WatermarkPattern`
- `InputDocument` 1 -> 1 `OutputDocument`
- `WatermarkPattern` influences each `PageImage.processing_mode`

## State Transitions

## ProcessingLifecycle

- `validated` -> `rasterized` -> `pattern_estimated` -> `pages_processed` -> `pdf_rebuilt` -> `completed`
- Failure transitions:
  - `validated` -> `failed` for invalid file or page count <2.
  - `pages_processed` continues with best-effort mode on low confidence pages; transitions to `failed` only on non-recoverable processing errors.
