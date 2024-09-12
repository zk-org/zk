# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information

project = "zk"
copyright = "2024, zk-org"
author = "zk-org"
release = "0.14.1"

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = ["myst_parser"]
myst_enable_extensions = ["colon_fence", "html_image"]
suppress_warnings = ["myst.xref_missing", "myst.iref_ambiguous"]

templates_path = ["_templates"]
exclude_patterns = [".zk"]


# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_theme = "furo"
html_static_path = ["_static"]
master_doc = "index"
