 ## First time running `gh tp`
 
 ```bash
 ➜  sausage-factory [main] [X] gh tp          
2025/02/28 22:56:56 Attention! Missing Config File: Config file should be named .tp.toml and exist in your home directory or in your project's root.
```
Let's grab our example config file and run `gh tp` again:

```bash
 ➜  sausage-factory [main] cp ../gh-tp/.tp.toml .                                                 
 ➜  sausage-factory [main] [X] gh tp
2025/02/28 22:57:17 plan.out was created
```

We can view the plan like so: 


  
```bash
 ➜  sausage-factory [main] [X] terraform show plan.out     
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

And we can verify the `plan.md` matches our output above :

```bash
 ➜  sausage-factory [main] [X] cat plan.md

```
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


We can then apply our Terraform:

```terraform
 ➜  sausage-factory [main] [X] terraform apply plan.out
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
 ➜  sausage-factory [main] [X] gh tp
2025/02/28 22:58:27 No changes.Your infrastructure matches the configuration.
2025/02/28 22:58:27 plan.out was created

```

We can verify that our Markdown contains as much:

```bash
 ➜  sausage-factory [main] [X] cat plan.md
 
   <details><summary>Terraform Plan</summary>

   ```terraform
   
   No Changes. Your Infrastructure matches the configuration.
   
   \```

   </details>

```

The `\` above exists to escape the code fences. That _does not_ exist in `gh tp` output. It purely exists for this presentation. If you haven't already seen it, view the sample markdown [example PR](./EXAMPLE-PR.md) to see a file created by `gh tp`.