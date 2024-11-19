variable "vultr_api_key" {
  type      = string
  default   = "${env("VULTR_API_KEY")}"
  sensitive = true
}

variable "vince_version" {
  type        = string
  default     = "${env("VINCE_VERSION")}"
  description = "Version number of the desired vince binary."
}

packer {
    required_plugins {
        vultr = {
            version = ">=v2.3.2"
            source = "github.com/vultr/vultr"
        }
    }
}

source "vultr" "vince" {
  api_key              = "${var.vultr_api_key}"
  os_id                = "387"
  plan_id              = "vc2-1c-1gb"
  region_id            = "ewr"
  snapshot_description = "vince-snapshot-${formatdate("YYYY-MM-DD hh:mm", timestamp())}"
  ssh_username         = "root"
  state_timeout        = "10m"
}

build {
  sources = ["source.vultr.vince"]

  provisioner "file" {
    source = "helper-scripts/vultr-helper.sh"
    destination = "/root/vultr-helper.sh"
  }

  provisioner "file" {
    source = "vince/setup-per-boot.sh"
    destination = "/root/setup-per-boot.sh"
  }

  # Copy configuration files
  provisioner "file" {
    destination = "/etc/"
    source      = "vince/etc/"
  }

  provisioner "file" {
    source = "vince/setup-per-instance.sh"
    destination = "/root/setup-per-instance.sh"
  }

  provisioner "shell" {
    environment_vars = [
      "VINCE_VERSION=${var.vince_version}",
      "DEBIAN_FRONTEND=noninteractive"
    ]
      script = "vince/vince.sh"
      remote_folder = "/root"
      remote_file = "vince.sh"
  }
}
