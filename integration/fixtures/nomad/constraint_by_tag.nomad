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
              "--providers.nomad.endpoint.address=http://127.0.0.3:4646",
              "--providers.nomad.constraints=Tag(`color=red`)",
            ]
          }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }

  group "who-red" {
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
        "color=red",
        "traefik.enable=true",
        "traefik.http.routers.example.entrypoints=web",
      ]
    }

    task "whoami" {
      driver = "docker"

      config {
        image = "traefik/whoami:v1.8.0"
        args  = ["-verbose", "-name", "whoami-red"]
      }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }

  group "who-blue" {
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
        "color=blue",
        "traefik.enable=true",
        "traefik.http.routers.example.entrypoints=web",
      ]
    }

    task "whoami" {
      driver = "docker"

      config {
        image = "traefik/whoami:v1.8.0"
        args  = ["-verbose", "-name", "whoami-blue"]
      }

      resources {
        cpu    = 10
        memory = 32
      }
    }
  }
}
