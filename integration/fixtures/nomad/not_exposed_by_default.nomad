job "who" {
  datacenters = ["dc1"]

  group "proxy" {
    network {
      mode = "host"
      port "ingress" {
        static = 8899
      }
    }

    task "traefik" {
      driver = "raw_exec"
      config {
        command = "EXECUTABLE"
        args = [
          "--log.level=DEBUG",
          "--entryPoints.web.address=:8899",
          "--providers=nomad",
          "--providers.nomad.refreshInterval=1s",
          "--providers.nomad.exposedByDefault=false",
          "--providers.nomad.endpoint.address=http://127.0.0.2:4646",
        ]
      }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }

  group "who-default" {
    network {
      mode = "bridge"
      port "http" {
        to = 80
      }
    }

    service {
      name     = "whoami"
      provider = "nomad"
      port     = "http"
      tags = [
        // disabled by default
        "traefik.http.routers.example.entrypoints=web",
      ]
    }

    task "whoami" {
      driver = "docker"

      config {
        image = "traefik/whoami:v1.8.0"
        args  = ["-verbose", "-name", "whoami-default"]
      }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }

  group "who-enable" {
    network {
      mode = "bridge"
      port "http" {
        to = 80
      }
    }

    service {
      name     = "whoami"
      provider = "nomad"
      port     = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.example.entrypoints=web",
      ]
    }

    task "whoami" {
      driver = "docker"

      config {
        image = "traefik/whoami:v1.8.0"
        args  = ["-verbose", "-name", "whoami-enabled"]
      }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }
}
