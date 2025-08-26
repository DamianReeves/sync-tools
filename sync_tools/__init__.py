__version__ = "0.1.0"

# Expose a package-level main() so the module can be used as an entrypoint
# for zipapps and python -m calls. This forwards to the CLI entrypoint.
from .__main__ import main  # noqa: E402,F401
