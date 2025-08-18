# Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
# SPDX-License-Identifier: MIT
#
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

"""Sphinx documentation configuration file."""

from datetime import datetime
import os

from ansys_sphinx_theme import ansys_logo_black, ansys_favicon, get_version_match

# Project information
project = "aali-flowkit"
copyright = f"(c) {datetime.now().year} ANSYS, Inc. All rights reserved"
author = "ANSYS, Inc."

# Read version from VERSION file
with open(os.path.join("..", "..", "VERSION"), "r") as f:
    version_file = f.readline().strip()

release = version = version_file
switcher_version = get_version_match(version_file)
cname = os.getenv("DOCUMENTATION_CNAME", "laughing-guide-5m1lvq6.pages.github.io")
"""The canonical name of the webpage hosting the documentation."""

# Select desired logo, theme, and declare the html title
html_theme = "ansys_sphinx_theme"
html_short_title = html_title = project
html_logo = ansys_logo_black
html_favicon = ansys_favicon
html_context = {
    "github_user": "ansys",
    "github_repo": "aali-flowkit",
    "github_version": "main",
    "doc_path": "doc/source",
}
html_theme_options = {
    "github_url": "https://github.com/ansys/aali-flowkit",
    "additional_breadcrumbs": [
        ("Aali", "https://aali.docs.ansys.com/"),
    ],
    "switcher": {
        "json_url": f"https://{cname}/versions.json",
        "version_match": switcher_version,
    },
    "check_switcher": False,
    "show_prev_next": False,
    "show_breadcrumbs": True,
    "use_edit_page_button": True,
}

# Sphinx extensions
extensions = [
    "sphinx_design",
    "sphinx_copybutton",
    "myst_parser",
]


# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

# The suffix(es) of source filenames.
source_suffix = ".rst"

# The master toctree document.
master_doc = "index"

source_suffix = {
    ".rst": "restructuredtext",
    ".md": "markdown",
}

# The master toctree document.
master_doc = "index"

# MyST parser configuration
myst_enable_extensions = [
    "replacements",
    "smartquotes",
]

# Enable heading anchors for proper fragment link resolution
myst_heading_anchors = 3

# Suppress cross-reference warnings for fragment links
suppress_warnings = ["myst.xref_missing"]
