`tp` is a GitHub [CLI](https://github.com/cli/cli) extension to create GitHub pull requests with [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/about-writing-and-formatting-on-github) containing the output from an [OpenTofu](https://opentofu.org/) or [Terraform](https://www.terraform.io/) plan's output [^1] [^2], wrapped around a [`<details></details>`](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-collapsed-sections) block so the plan output can be collapsed for easier reading. The body of your pull request will look like this [example](./example/EXAMPLE-PR.md) in the example directory.

> [!TIP]
> View it in 'rich diff' mode to see the rendered view.

## Getting Started | Installation

```bash
gh ext install esacteksab/gh-tp
```

### `.tp.toml` config file

I wanted to make as few assumptions about your environment as possible, so `tp` does not define any default values today. `tp` uses a config file named `.tp.toml`. This config file is written in [TOML](https://toml.io/). TOML is case-sensitive and keys are [mixedCase or camelCase](https://en.wikipedia.org/wiki/Camel_case) where applicable. It has 2 required parameters with one optional parameters. The config file is expected to exist in either your home directory or in the root of your project from where you will execute `gh tp` from. A annotated copy exists in the [example](./example) directory. **_The config file, the parameters and possibly the presence of default values is actively being worked on. This behavior may change in a future release._**

| Parameter | Type   | Required | Description                                                                                                                      |
| --------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| binary    | string | N        | We look on your `$PATH` for `tofu` or `terraform`, if both exist, you _must_ define _one_ in your config. _Default: `undefined`_ |
| planFile  | string | Y        | The name of the plan's output file created by `gh tp`. _Default: `""`_                                                           |
| mdFile    | string | Y        | The name of the Markdown file created by `gh tp`. _Default: `""`_                                                                |

### Using `tp`

To create a plan and the markdown from that plan, run

```bash
gh tp
Using config file: /Users/tempuser/.tp.toml
```

Two files will be created, the first an output file named, what you defined for the value of `planFile` in `.tp` config and a Markdown file named what you defined for the value of the parameter `mdFile` in the `.tp` config file.

```bash
ls |grep plan
plan.md
plan.out
```

### Create Commit

```bash
git add .
git commit -m "feat: adding cool things"
```

### Create Pull Request with `gh`

While the end goal of this extension is to submit the pull request, that functionality doesn't exist on a public branch yet. **This feature exists in a prototype branch. It will be public _soon_!** So to submit a pull request today, we can use the built-in functionality of `gh`.

```bash
gh pr create -F plan.md
... gh things occur here ...
https://github.com/org/repo/pull/42
```

### View Pull Request in Browser with `gh`

```bash
gh pr view -w
Opening https://github.com/org/repo/pull/42 in your browser.
```

### Targeting with Terraform

If you're targeting a resource, you can still create markdown from that plan's output. `tp` reads from `stdin` like so:

```bash
terraform plan -out plan.out -no-color  | gh tp -
```

Like with `gh tp` two files will exist. The first being whatever you passed to `-out` for the file name in the above example (`plan.out` in the example above) and the Markdown file named whatever you defined as the value for the `mdFile` parameter in the `.tp` config file. `tp` does not create an additional plan having been passed the plan from `stdin`.

### Extended Example

The above example is intended to be just enough to get you started. If you'd like to see an example representative of a more real-world use case, one exists in the [example](./example/) directory. A note though, I've been unable to figure out how to put Markdown with code fences inside Markdown code fences. So the formatting on that example exists purely out of a need to handle the situation where I output Markdown and I'm trying to put it inside code fences. I hope you understand and I hope I can come up with a solution long-term to better display the output of `tp`.

### Disable Terminal Colors

> [!TIP]
> To disable color output in the terminal, set the environment variable `NO_COLOR` to `true`

<!-- markdownlint-disable-line MD028 -->

> [!WARNING]
> **_Pre 1.0.0 Alert_**. This is a new extension and as a result, the API isn't set yet. There have been two prototypes so far. The experiences from both have shaped what exists today, but there is still quite a bit left to do. So as I continue down the path of figuring things out, things are almost certainly going to change. **_Expect_** breaking changes in the early releases. I will _strive_ not to publish broken builds and will lean into Semver to identify `alpha`, `beta` and pre-releases.

## Motivation

I write _a lot_ of Terraform daily and a part of that process includes submitting pull requests for review by peers prior to applying the plan and merging the pull request. It can be tedious, sometimes cumbersome to run a `terraform plan -out plan.out`, copy the output from the terminal, do Git "things", open a pull request, paste the contents of the plan's output into the body of the pull request, wrap it in a code block, then wrap _that_ in a `<details></details>` block because some Terraform output can be quite lengthy and in an effort to provide a better experience and quicker access to the pull request's comments, we use the `<details></details>` mechanism to collapse the plan's output. You can use [pull request Templates](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests) to short-circuit some of this workflow with the template already containing a pre-formatted code block wrapped in a `<details></details>` so all you have to do is paste your plan in the WYSIWYG editor and hit submit.

~~But I'm _lazy_~~, I mean, I'm _intentional_ with the things that take my time and attention, so I dug into letting the robots do what they do best and `tp` was born! The first prototype of this was a shell script and some `sed` and `awk` leveraging existing functionality present in `gh`. I decided to continue to iterate and extend it further by writing this extension in Go.

## What tp isn't

I feel like I'm in this _weird_ space. I programmatically run a `terraform plan -out plan.out --no-color` but Terraform _already_ does that. And it's not my intent to create a wrapper around an existing tool, especially one like Terraform. I also programmatically do a `gh pr create -t $title -F file.md`, but `gh` _already_ does that. So while I find my fit in the space, I felt it was important to call out what I'm not going to do. Today, it's not uncommon for me to have to do a `-target` to plan/apply around something. `tp` doesn't natively support passing arguments to Terraform. And I don't think I want it to. So in the example of not being able to pass the `-target` argument, but still desiring to create the formatted Markdown and the subsequent pull request, `tp` can read from `stdin` so today you can run `terraform plan -target resource.name -out plan.out | gh tp -` and `tp` will create the Markdown with your plan's output.

<!--## Contribute

### Local Development Setup-->

#### Disclaimer

> [!NOTE]
> This is a personal project that was born out of need and want to automate the repetitive task out of my life. `tp` is in no way affiliated with or associated with Terraform, HashiCorp, OpenTofu or any entities official or unofficial. The views expressed here are my own and don't reflect any past, current or future employer.

[^1]: https://opentofu.org/docs/cli/commands/plan/#other-options <!-- markdownlint-disable-line MD034 -->

[^2]: https://developer.hashicorp.com/terraform/cli/commands/plan#out-filename <!-- markdownlint-disable-line MD034 -->
