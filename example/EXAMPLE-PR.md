<!-- markdownlint-disable MD033 -->

<details><summary>Terraform Plan</summary>

```terraform

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
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
