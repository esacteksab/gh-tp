<!-- markdownlint-disable MD033 -->

## First time running `gh tp`

```bash
gh tp
2025/03/04 02:08:12 ERRO Missing Config File: Config file should be named .tp.toml and exist in your home directory or in your project's root.
```

Let's grab our example config file and run `gh tp` again:

```bash
cp ../gh-tp/.tp.toml .
gh tp
✔  Plan Created...
✔  Markdown Created...
```

> [!NOTE]
> On projects with a large amount of resources, creating the plan can take some time. Currently `tp` does not provided feedback that it's doing anything. This might create the situation where you're wondering if it is doing anything, and contemplate pressing `CTRL-C`. Providing this awareness that things are happening is being tracked in this issue [Feat: Is it Doing Anything?](https://github.com/esacteksab/gh-tp/issues/20).

### We can view the plan like so

```bash
terraform show plan.out
```

<details>

```terraform

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # archive_file.tf_pr will be created
  + resource "archive_file" "tf_pr" {
      + id                  = (known after apply)
      + output_base64sha256 = (known after apply)
      + output_base64sha512 = (known after apply)
      + output_md5          = (known after apply)
      + output_path         = "./tf-pr.tar.gz"
      + output_sha          = (known after apply)
      + output_sha256       = (known after apply)
      + output_sha512       = (known after apply)
      + output_size         = (known after apply)
      + source_file         = "./.tf-pr"
      + type                = "tar.gz"
    }

  # local_file.pet will be created
  + resource "local_file" "pet" {
      + content              = (known after apply)
      + content_base64sha256 = (known after apply)
      + content_base64sha512 = (known after apply)
      + content_md5          = (known after apply)
      + content_sha1         = (known after apply)
      + content_sha256       = (known after apply)
      + content_sha512       = (known after apply)
      + directory_permission = "0777"
      + file_permission      = "0777"
      + filename             = "./pet.out"
      + id                   = (known after apply)
    }

  # local_file.uuid will be created
  + resource "local_file" "uuid" {
      + content              = (known after apply)
      + content_base64sha256 = (known after apply)
      + content_base64sha512 = (known after apply)
      + content_md5          = (known after apply)
      + content_sha1         = (known after apply)
      + content_sha256       = (known after apply)
      + content_sha512       = (known after apply)
      + directory_permission = "0777"
      + file_permission      = "0777"
      + filename             = "./uuid.out"
      + id                   = (known after apply)
    }

  # random_pet.pet will be created
  + resource "random_pet" "pet" {
      + id        = (known after apply)
      + length    = 2
      + separator = "-"
    }

  # random_uuid.uuid will be created
  + resource "random_uuid" "uuid" {
      + id     = (known after apply)
      + result = (known after apply)
    }

Plan: 5 to add, 0 to change, 0 to destroy.
```

</details>

And we can verify the `plan.md` matches our output above:

```bash
cat plan.md
```

````md
<details><summary>Terraform Plan</summary>

\```terraform

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:

- create

Terraform will perform the following actions:

# archive_file.tf_pr will be created

- resource "archive_file" "tf_pr" {
  - id = (known after apply)
  - output_base64sha256 = (known after apply)
  - output_base64sha512 = (known after apply)
  - output_md5 = (known after apply)
  - output_path = "./tf-pr.tar.gz"
  - output_sha = (known after apply)
  - output_sha256 = (known after apply)
  - output_sha512 = (known after apply)
  - output_size = (known after apply)
  - source_file = "./.tf-pr"
  - type = "tar.gz"
    }

# local_file.pet will be created

- resource "local_file" "pet" {
  - content = (known after apply)
  - content_base64sha256 = (known after apply)
  - content_base64sha512 = (known after apply)
  - content_md5 = (known after apply)
  - content_sha1 = (known after apply)
  - content_sha256 = (known after apply)
  - content_sha512 = (known after apply)
  - directory_permission = "0777"
  - file_permission = "0777"
  - filename = "./pet.out"
  - id = (known after apply)
    }

# local_file.uuid will be created

- resource "local_file" "uuid" {
  - content = (known after apply)
  - content_base64sha256 = (known after apply)
  - content_base64sha512 = (known after apply)
  - content_md5 = (known after apply)
  - content_sha1 = (known after apply)
  - content_sha256 = (known after apply)
  - content_sha512 = (known after apply)
  - directory_permission = "0777"
  - file_permission = "0777"
  - filename = "./uuid.out"
  - id = (known after apply)
    }

# random_pet.pet will be created

- resource "random_pet" "pet" {
  - id = (known after apply)
  - length = 2
  - separator = "-"
    }

# random_uuid.uuid will be created

- resource "random_uuid" "uuid" {
  - id = (known after apply)
  - result = (known after apply)
    }

Plan: 5 to add, 0 to change, 0 to destroy.
\```

</details>
````

### We can then apply our Terraform

```terraform
terraform apply plan.out
random_uuid.uuid: Creating...
random_pet.pet: Creating...
archive_file.tf_pr: Creating...
random_pet.pet: Creation complete after 0s [id=notable-chimp]
random_uuid.uuid: Creation complete after 0s [id=3716c1b9-746d-bb77-88c6-9559293517d8]
archive_file.tf_pr: Creation complete after 0s [id=2393566a4ef1b417793d52c8f119147fce053b25]
local_file.pet: Creating...
local_file.uuid: Creating...
local_file.pet: Creation complete after 0s [id=8c0a7420e3c6cf9c3e3b39047c5e4688c6252cae]
local_file.uuid: Creation complete after 0s [id=05eefd70eda2775e876d248a874c4e8e84ba8c0d]

Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
```

Run `gh tp` again, and there are no changes as expected.

```bash
gh tp
✔  Plan Created...
✔  Markdown Created...
```

We can verify that our Markdown contains as much:

```bash
cat plan.md
```

````md
<details><summary>Terraform Plan</summary>

\```terraform

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.

\```

</details>
````

> [!NOTE]
> The `\` above exists to escape the code fences. That _does not_ exist in `gh tp` output. It purely exists for this presentation. If you haven't already seen it, view the sample markdown [example PR](./EXAMPLE-PR.md) to see a file created by `gh tp`.
