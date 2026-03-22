// ==================================================
// Store Definition
// ==================================================

store "homelab" {
  namespace = "homelab"

  labels = {
    component   = "storage"
    environment = "homelab"
  }

  backend = "local"

  paths {
    artifacts = "/home/zakaria/dox/homelab/artifacts"
    images    = "/home/zakaria/dox/homelab/images"
  }

  image "rocky-9.5" {
    display    = "Rocky Linux 9.5"
    version    = "9.5"
    os_profile = "https://rockylinux.org/rocky/9"
    file       = "rocky/rocky-9-5-base-image.qcow2"
    size       = "2.6G"
    checksum   = "sha256:eedbdc2875c32c7f00e70fc861edef48587c7cbfd106885af80bdf434543820b"
  }

  image "rocky-10.1" {
    display    = "Rocky Linux 10.1"
    version    = "10"
    os_profile = "https://rockylinux.org/rocky/10"
    file       = "rocky/rocky-10-1-base-image.qcow2"
    size       = "520M"
    checksum   = "sha256:eedbdc2875c32c7f00e70fc861edef48587c7cbfd106885af80bdf434543820b"
  }

}
