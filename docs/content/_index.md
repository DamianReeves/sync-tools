---
title: "sync-tools"
linkTitle: "sync-tools"
---

# sync-tools — Fast directory sync with Go, Cobra, and Bubble Tea

{{< blocks/cover title="sync-tools" image_anchor="top" height="full" color="orange" >}}
<div class="mx-auto">
	<a class="btn btn-lg btn-primary mr-3 mb-4" href="{{< relref "/docs" >}}">
		Learn More <i class="fas fa-arrow-alt-circle-right ml-2"></i>
	</a>
	<a class="btn btn-lg btn-secondary mr-3 mb-4" href="https://github.com/DamianReeves/sync-tools/releases">
		Download <i class="fab fa-github ml-2 "></i>
	</a>
	<p class="lead mt-5">Fast & efficient directory synchronization with Git patch support!</p>
	{{< blocks/link-down color="info" >}}
</div>
{{< /blocks/cover >}}

{{% blocks/lead color="primary" %}}
sync-tools is a powerful, modern Go CLI wrapper around rsync that provides fast directory synchronization with advanced features like Git patch generation, interactive preview, and declarative SyncFile configuration.
{{% /blocks/lead %}}

{{< blocks/section color="dark" >}}
{{% blocks/feature icon="fa-lightbulb" title="Fast & Efficient" %}}
Built with Go for high performance and cross-platform support. Uses rsync for efficient file transfers.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-arrows-alt" title="One-way or Two-way Sync" %}}
Support for both one-way and two-way directory synchronization with conflict detection.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-file-alt" title="Git Patch Generation" %}}
Generate git-format patch files instead of syncing for review and manual application workflows.
{{% /blocks/feature %}}

{{< /blocks/section >}}

{{< blocks/section >}}

<div class="col">
<h2 class="text-center">Key Features</h2>

- **🚀 Fast & Efficient**: Built with Go for high performance and cross-platform support
- **🎯 One-way or two-way** directory synchronization
- **📁 Gitignore-style** `.syncignore` files (source and destination)
- **🔗 Optional import** of `SOURCE/.gitignore` patterns
- **🎨 Interactive Mode**: Beautiful terminal UI with Bubble Tea
- **📜 SyncFile Format**: Dockerfile-like declarative sync configuration
- **⚡ Per-side ignore** files and inline patterns (with `!` unignore)
- **📋 "Whitelist" mode** to sync only specified paths
- **⚙️ Flexible Configuration**: TOML config files OR pure CLI usage
- **🔍 Smart Defaults**: Excludes `.git/`, optional hidden directory exclusion
- **🎭 Dry-run previews** and detailed change output
- **📊 Multiple Output Formats**: Text, JSON logging, Markdown reports, and git patches
- **🔧 Git Patch Generation**: Create git-format patch files instead of syncing (via --patch flag or --report with .patch/.diff extension)
- **👁 Preview Mode**: Show colored diff preview with paging support
- **✅ Apply Patches**: Generate and apply patches with confirmation prompts

</div>

{{< /blocks/section >}}

{{< blocks/section color="primary" >}}
<div class="col-12">
<h2 class="text-center">Ready to get started?</h2>
<p class="text-center">
<a class="btn btn-lg btn-primary mr-3 mb-4" href="{{< relref "/docs/getting-started" >}}">Get Started</a>
<a class="btn btn-lg btn-secondary mr-3 mb-4" href="{{< relref "/docs/examples" >}}">Examples</a>
</p>
</div>
{{< /blocks/section >}}
