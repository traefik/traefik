namespace API {
  type Version = {
    Codename: string
    Version: string
    disableDashboardAd: boolean
    startDate: string
  }

  type Certificate = {
    name: string
    expiration: string
    domains: string[]
  }
}
