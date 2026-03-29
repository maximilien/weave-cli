#!/usr/bin/env python3
"""Generate test PDFs for weave-cli PDF extraction tests (Issue #8)."""

import os
import random
from fpdf import FPDF
from PIL import Image

BASE = "/tmp/weave-cli/tests/fixtures"
MC = dict(new_x="LMARGIN", new_y="NEXT")  # multi_cell kwargs


def make_dirs():
    os.makedirs(f"{BASE}/pdf_versions", exist_ok=True)
    os.makedirs(f"{BASE}/pdf_types", exist_ok=True)


LOREM = (
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. "
    "Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
    "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris "
    "nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in "
    "reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla "
    "pariatur. Excepteur sint occaecat cupidatat non proident, sunt in "
    "culpa qui officia deserunt mollit anim id est laborum."
)

TECH_PARAGRAPHS = [
    "Weave-CLI PDF Extraction Test Document. This document tests PDF text extraction across different PDF versions and types.",
    "Section 1: Basic Text. The quick brown fox jumps over the lazy dog. This sentence contains every letter of the English alphabet and is commonly used for font testing.",
    "Section 2: Technical Content. Vector databases store high-dimensional embeddings for similarity search. Common operations include insert, query, delete, and update. Supported backends include Weaviate, Milvus, Chroma, Qdrant, OpenSearch, and Neo4j.",
    "Section 3: Data Processing. PDF extraction involves parsing the document structure, extracting text streams, identifying image objects, and reconstructing the reading order. Chunk sizes affect downstream embedding quality.",
    "Section 4: Quality Assurance. Automated testing validates that extraction produces consistent, complete results across PDF versions. Metadata preservation ensures traceability back to source documents.",
]


def add_standard_content(pdf):
    """Add standard text content to a PDF."""
    for para in TECH_PARAGRAPHS:
        pdf.multi_cell(w=0, h=6, text=para, **MC)
        pdf.ln(4)
    for i in range(3):
        pdf.multi_cell(w=0, h=6, text=f"Extended paragraph {i+1}: {LOREM}", **MC)
        pdf.ln(3)


def gen_versioned_pdf(version_str, filename):
    """Generate a PDF and patch the version header."""
    pdf = FPDF()
    pdf.set_auto_page_break(auto=True, margin=15)
    pdf.add_page()
    pdf.set_font("Helvetica", "B", 16)
    pdf.cell(0, 10, f"PDF Version {version_str} Test Document", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.ln(5)
    add_standard_content(pdf)

    path = f"{BASE}/pdf_versions/{filename}"
    pdf.output(path)

    # Patch the PDF header to claim the target version
    with open(path, "r+b") as f:
        header = f.read(20)
        idx = header.find(b"%PDF-")
        if idx >= 0:
            f.seek(idx)
            f.write(f"%PDF-{version_str}".encode())

    print(f"  Created {filename} (PDF {version_str})")


def gen_text_only_pdf():
    """Generate a text-only PDF (no images)."""
    pdf = FPDF()
    pdf.set_auto_page_break(auto=True, margin=15)
    pdf.add_page()
    pdf.set_font("Helvetica", "B", 18)
    pdf.cell(0, 12, "Text-Only PDF Test Document", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.ln(5)
    add_standard_content(pdf)

    for i in range(8):
        pdf.set_font("Helvetica", "B", 12)
        pdf.cell(0, 8, f"Chapter {i+1}: Extended Content", **MC)
        pdf.set_font("Helvetica", "", 11)
        pdf.multi_cell(w=0, h=6, text=LOREM, **MC)
        pdf.multi_cell(w=0, h=6, text=LOREM, **MC)
        pdf.ln(3)

    path = f"{BASE}/pdf_types/text_only.pdf"
    pdf.output(path)
    print(f"  Created text_only.pdf")


def create_test_image(path, width=200, height=150, color=(70, 130, 180)):
    """Create a simple test image."""
    img = Image.new("RGB", (width, height), color)
    pixels = img.load()
    for x in range(0, width, 20):
        for y in range(height):
            pixels[x, y] = (255, 255, 255)
    for y in range(0, height, 20):
        for x in range(width):
            pixels[x, y] = (200, 200, 200)
    img.save(path)
    return path


def gen_mixed_pdf():
    """Generate a PDF with both text and images."""
    pdf = FPDF()
    pdf.set_auto_page_break(auto=True, margin=15)

    img1 = create_test_image("/tmp/test_img1.png", 300, 200, (70, 130, 180))
    img2 = create_test_image("/tmp/test_img2.png", 250, 180, (180, 70, 70))

    pdf.add_page()
    pdf.set_font("Helvetica", "B", 18)
    pdf.cell(0, 12, "Mixed Content PDF Test Document", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.ln(5)
    pdf.multi_cell(w=0, h=6, text="This document contains both text and images for extraction testing.", **MC)
    pdf.ln(5)

    pdf.set_font("Helvetica", "B", 14)
    pdf.cell(0, 10, "Section 1: Text with Inline Image", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.multi_cell(w=0, h=6, text=LOREM, **MC)
    pdf.ln(3)
    pdf.image(img1, x=30, w=100)
    pdf.ln(5)
    pdf.multi_cell(w=0, h=6, text="The image above shows a test pattern used for extraction validation.", **MC)
    pdf.ln(5)

    pdf.set_font("Helvetica", "B", 14)
    pdf.cell(0, 10, "Section 2: More Content", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.multi_cell(w=0, h=6, text=LOREM, **MC)
    pdf.ln(3)
    pdf.image(img2, x=30, w=100)
    pdf.ln(5)
    pdf.multi_cell(w=0, h=6, text="Second test image above. Both images should be extractable.", **MC)

    for i in range(5):
        pdf.ln(3)
        pdf.multi_cell(w=0, h=6, text=f"Additional paragraph {i+1}: {LOREM}", **MC)

    path = f"{BASE}/pdf_types/mixed.pdf"
    pdf.output(path)
    print(f"  Created mixed.pdf")


def gen_photo_heavy_pdf():
    """Generate a PDF with many images and some text."""
    pdf = FPDF()
    pdf.set_auto_page_break(auto=True, margin=15)

    colors = [
        (70, 130, 180), (180, 70, 70), (70, 180, 70),
        (180, 180, 70), (180, 70, 180), (70, 180, 180),
        (120, 80, 40), (40, 80, 120),
    ]

    pdf.add_page()
    pdf.set_font("Helvetica", "B", 18)
    pdf.cell(0, 12, "Photo-Heavy PDF Test Document", **MC)
    pdf.set_font("Helvetica", "", 11)
    pdf.ln(3)
    pdf.multi_cell(w=0, h=6, text="This document is image-heavy with 8 test images across multiple pages.", **MC)

    for i, color in enumerate(colors):
        img_path = f"/tmp/test_heavy_{i}.png"
        create_test_image(img_path, 400, 250, color)
        pdf.ln(3)
        pdf.set_font("Helvetica", "B", 12)
        pdf.cell(0, 8, f"Image {i+1}: Test Pattern RGB({color[0]},{color[1]},{color[2]})", **MC)
        pdf.set_font("Helvetica", "", 11)
        pdf.multi_cell(w=0, h=6, text=f"Description for image {i+1} with color pattern.", **MC)
        pdf.ln(2)
        pdf.image(img_path, x=20, w=120)

    path = f"{BASE}/pdf_types/photo_heavy.pdf"
    pdf.output(path)
    print(f"  Created photo_heavy.pdf")


def gen_scanned_pdf():
    """Generate a PDF that simulates a scanned document (image-only, no text layer)."""
    random.seed(42)

    pdf = FPDF()
    for page_num in range(2):
        bg = (245, 240, 230) if page_num == 0 else (240, 238, 228)
        img = Image.new("RGB", (612, 792), bg)
        pixels = img.load()
        for _ in range(5000):
            x = random.randint(0, 611)
            y = random.randint(0, 791)
            gray = random.randint(200, 240)
            pixels[x, y] = (gray, gray, gray)

        scan_path = f"/tmp/scanned_page_{page_num}.png"
        img.save(scan_path)
        pdf.add_page()
        pdf.image(scan_path, x=0, y=0, w=210, h=297)

    path = f"{BASE}/pdf_types/scanned.pdf"
    pdf.output(path)
    print(f"  Created scanned.pdf")


if __name__ == "__main__":
    make_dirs()

    print("Generating PDF version fixtures:")
    gen_versioned_pdf("1.3", "pdf_1.3.pdf")
    gen_versioned_pdf("1.4", "pdf_1.4.pdf")
    gen_versioned_pdf("1.7", "pdf_1.7.pdf")
    gen_versioned_pdf("2.0", "pdf_2.0.pdf")

    print("\nGenerating PDF type fixtures:")
    gen_text_only_pdf()
    gen_mixed_pdf()
    gen_photo_heavy_pdf()
    gen_scanned_pdf()

    print("\nDone!")
