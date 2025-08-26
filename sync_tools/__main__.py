"""Module entrypoint so the package can be run with -m or as a zipapp.

This simply forwards into the CLI entrypoint.
"""
from .cli import cli


def main() -> None:
    # Call the click CLI. When executed as a zipapp the program name will be
    # set by the zipapp shebang or the user invocation.
    cli()


if __name__ == "__main__":
    main()
