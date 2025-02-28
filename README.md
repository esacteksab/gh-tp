`tp` is a GitHub [CLI](https://github.com/cli/cli) extension to create GitHub pull requests with [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/about-writing-and-formatting-on-github) containing the output from an [OpenTofu](https://opentofu.org/) or [Terraform](https://www.terraform.io/) plan's output [^1] [^2], wrapped around a [`<details></details>`](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-collapsed-sections) block so the plan output can be collapsed for easier reading. The body of your pull request will look like this [example](./example/EXAMPLE-PR.md) in the example directory.

> [!TIP]
> View it in 'rich diff' mode to see the rendered view.

## Getting Started | Installation

```bash
gh ext install esacteksab/gh-tp
```

### `.tp` config file

I didn't want to make assumptions about your system, so `tp` does not define any default values today. `tp` uses a config file named `.tp`. This config file is written in [YAML](https://yaml.org/). The file today is rudimentary. It has 3 required parameters with one optional parameter. It can exist in your home directory or in the root of your project from where you will execute `gh tp` from. A annotated copy exists in the [example](./example) directory. **_The config file, the parameters and possibly the presence of default values is actively being worked on. This behavior may change in a future release._**

| Parameter | Type   | Required | Description                                                                                                            |
| --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------- |
| binary    | string | Y        | The name of the binary that you use. (e.g. `tofu` or `terraform`). It is expected to be in your $PATH. _Default: `""`_ |
| planfile  | string | Y        | The name of the plan output file. _Default: `""`_                                                                      |
| mdfile    | string | Y        | The name of the markdown file. _Default: `""`_                                                                         |
| no-color  | bool   | N        | If `true`, `tp` will emit no color on the terminal. _Default: `false`_                                                 |

### Using `tp`

To create a plan and the markdown from that plan, run

```bash
gh tp
```

Two files will be created, the first an output file named, what you defined for the value of `planfile` in `.tp` config and a Markdown file named what you defined for the value of the parameter `mdfile` in the `.tp` config file. While the end goal of this extension is to submit the pull request, that functionality doesn't exist on a public branch yet. **This feature exists in a prototype branch. It will be public _soon_!**

If you're targeting a resource, you can still get markdown from that plan's output. `tp` reads from `stdin` like so:

```bash
terraform plan -out plan.out -no-color  | gh tp -
```

Like with `gh tp` two files will exist. The first being whatever you passed to `-out` for the file name in the above example (`plan.out` in the example above) and the Markdown file named whatever you defined as the value for the `mdfile` parameter in the `.tp` config file. `tp` does not create an additional plan having been passed the plan from `stdin`.

<!--`tp` also supports command-line flags as well as source environment variables. More [below](#disclaimer)-->

> [!WARNING]
> **_Pre 1.0.0 Alert_**. This is a new extension and as a result, the API isn't set yet. There have been two prototypes so far. The experiences from both have shaped what exists today, but there is still quite a bit left to do. So as I continue down the path of figuring things out, things are almost certainly going to change. **_Expect_** breaking changes in the early releases. I will _strive_ not to publish broken builds and will lean into Semver to identify `alpha`, `beta` and pre-releases.

## Motivation

I write _a lot_ of Terraform daily and a part of that process includes submitting pull requests for review by peers prior to applying the plan and merging the pull request. It can be tedious, sometimes cumbersome to run a `terraform plan -out plan.out`, copy the output from the terminal, do Git "things", open a pull request, paste the contents of the plan's output into the body of the pull request, wrap it in a code block, then wrap _that_ in a `<details></details>` block because some Terraform output can be quite lengthy and in an effort to provide a better experience and quicker access to the pull request's comments, we use the `<details></details>` mechanism to collapse the plan's output. You can use [pull request Templates](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests) to short-circuit some of this workflow with the template already containing a pre-formatted code block wrapped in a `<details></details>` so all you have to do is paste your plan in the WYSIWYG editor and hit submit.

~~But I'm _lazy_~~, I mean, I'm _intentional_ with the things that take my time and attention, so I dug into letting the robots do what they do best and `tp` was born! The first prototype of this was a shell script and some `sed` and `awk` leveraging existing functionality present in `gh`. I decided to continue to iterate and extend it further by writing this extension in Go.

## What tp isn't

I feel like I'm in this _weird_ space. I programmatically run a `terraform plan -out plan.out --no-color` but Terraform _already_ does that. And it's not my intent to create a wrapper around an existing tool, especially one like Terraform. I also programmatically do a `gh pr create -t $title -F file.md`, but `gh` _already_ does that. So while I find my fit in the space, I felt it was important to call out what I'm not going to do. Today, it's not uncommon for me to have to do a `-target` to plan/apply around something. `tp` doesn't natively support passing arguements to Terraform. And I don't think I want it to. So in the example of not being able to pass the `-target` argument, but still desiring to create the formatted Markdown and the subsequent pull request, `tp` can read from `stdin` so today you can run `terraform plan -target resource.name -out plan.out | gh tp -` and `tp` will create the Markdown with your plan's output.

## Contribute

### Local Development Setup

#### Disclaimer

> [!NOTE]
> This is a personal project that was born out of need and want to automate the repetitive task out of my life. `tp` is in no way affiliated with or associated with Terraform, HashiCorp, OpenTofu or any entities official or unofficial. The views expressed here are my own and don't reflect any past, current or future employer.

<!--I'm left feeling like "Create the targeted plan with Terraform and let `tp` do the rest! While doing early prototyping, leaning into HashiCorp's example of how to use the `terraform-exec` library, I had in a `tf.Init` and while iterating, I kept doing an init and it kept downloading providers. I'm pretty certain I got rate-limited by the registry. So do I allow a `-i --init` to be passed so folks can do it when they need to or do I jump back on the side of "Have Terraform do the init, come back to `tp` when you're ready!"?-->

[^1]: https://opentofu.org/docs/cli/commands/plan/#other-options <!-- markdownlint-disable-line MD034 -->

[^2]: https://developer.hashicorp.com/terraform/cli/commands/plan#out-filename <!-- markdownlint-disable-line MD034 -->
