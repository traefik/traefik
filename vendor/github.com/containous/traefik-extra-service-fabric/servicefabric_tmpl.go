package servicefabric

const tmpl = `
[backends]
    {{$groupedServiceMap := getServicesWithLabelValueMap .Services "backend.group.name"}}
    {{range $aggName, $aggServices := $groupedServiceMap }}
      [backends."{{$aggName}}"]
      {{range $service := $aggServices}}
        {{range $partition := $service.Partitions}}
          {{range $instance := $partition.Instances}}
            [backends."{{$aggName}}".servers."{{$service.ID}}-{{$instance.ID}}"]
            url = "{{getDefaultEndpoint $instance}}"
            weight = {{getServiceLabelValueWithDefault $service "backend.group.weight" "1"}}
          {{end}}
        {{end}}
      {{end}}
    {{end}}
  {{range $service := .Services}}
    {{range $partition := $service.Partitions}}
      {{if eq $partition.ServiceKind "Stateless"}}
      [backends."{{$service.Name}}"]
        [backends."{{$service.Name}}".LoadBalancer]
        {{if hasServiceLabel $service "backend.loadbalancer.method"}}
          method = "{{getServiceLabelValue $service "backend.loadbalancer.method" }}"
        {{else}}
          method = "drr"
        {{end}}

        {{if hasServiceLabel $service "backend.healthcheck"}}
          [backends."{{$service.Name}}".healthcheck]
          path = "{{getServiceLabelValue $service "backend.healthcheck"}}"
          interval = "{{getServiceLabelValueWithDefault $service "backend.healthcheck.interval" "10s"}}"
        {{end}}

        {{if hasServiceLabel $service "backend.loadbalancer.stickiness"}}
          [backends."{{$service.Name}}".LoadBalancer.stickiness]
        {{end}}

        {{if hasServiceLabel $service "backend.circuitbreaker"}}
          [backends."{{$service.Name}}".circuitbreaker]
          expression = "{{getServiceLabelValue $service "backend.circuitbreaker"}}"
        {{end}}

        {{if hasServiceLabel $service "backend.maxconn.amount"}}
          [backends."{{$service.Name}}".maxconn]
          amount = {{getServiceLabelValue $service "backend.maxconn.amount"}}
          {{if hasServiceLabel $service "backend.maxconn.extractorfunc"}}
          extractorfunc = "{{getServiceLabelValue $service "backend.maxconn.extractorfunc"}}"
          {{end}}
        {{end}}

        {{range $instance := $partition.Instances}}
          [backends."{{$service.Name}}".servers."{{$instance.ID}}"]
          url = "{{getDefaultEndpoint $instance}}"
          weight = {{getServiceLabelValueWithDefault $service "backend.weight" "1"}}
        {{end}}
      {{else if eq $partition.ServiceKind "Stateful"}}
        {{range $replica := $partition.Replicas}}
          {{if isPrimary $replica}}

            {{$backendName := (print $service.Name $partition.PartitionInformation.ID)}}
            [backends."{{$backendName}}".servers."{{$replica.ID}}"]
            url = "{{getDefaultEndpoint $replica}}"
            weight = 1

            [backends."{{$backendName}}".LoadBalancer]
            method = "drr"

            [backends."{{$backendName}}".circuitbreaker]
            expression = "NetworkErrorRatio() > 0.5"

          {{end}}
        {{end}}
      {{end}}
    {{end}}
{{end}}

[frontends]
{{range $groupName, $groupServices := $groupedServiceMap}}
  {{$service := index $groupServices 0}}
    [frontends."{{$groupName}}"]
    backend = "{{$groupName}}"

    {{if hasServiceLabel $service "frontend.priority"}}
    priority = 100
    {{end}}

    {{range $key, $value := getServiceLabelsWithPrefix $service "frontend.rule"}}
    [frontends."{{$groupName}}".routes."{{$key}}"]
    rule = "{{$value}}"
    {{end}}
{{end}}
{{range $service := .Services}}
  {{if isExposed $service}}
    {{if eq $service.ServiceKind "Stateless"}}

    [frontends."{{$service.Name}}"]
    backend = "{{$service.Name}}"

    {{if hasServiceLabel $service "frontend.passHostHeader"}}
      passHostHeader = {{getServiceLabelValue $service "frontend.passHostHeader" }}
    {{end}}

    {{if hasServiceLabel $service "frontend.whitelistSourceRange"}}
      whitelistSourceRange = {{getServiceLabelValue $service "frontend.whitelistSourceRange" }}
    {{end}}

    {{if hasServiceLabel $service "frontend.priority"}}
      priority = {{getServiceLabelValue $service "frontend.priority"}}
    {{end}}

    {{if hasServiceLabel $service "frontend.basicAuth"}}
      basicAuth = {{getServiceLabelValue $service "frontend.basicAuth"}}
    {{end}}

    {{if hasServiceLabel $service "frontend.entryPoints"}}
      entryPoints = {{getServiceLabelValue $service "frontend.entryPoints"}}
    {{end}}

    {{range $key, $value := getServiceLabelsWithPrefix $service "frontend.rule"}}
    [frontends."{{$service.Name}}".routes."{{$key}}"]
    rule = "{{$value}}"
    {{end}}

    {{else if eq $service.ServiceKind "Stateful"}}
      {{range $partition := $service.Partitions}}
        {{$partitionId := $partition.PartitionInformation.ID}}

        {{if hasServiceLabel $service "frontend.rule"}}
          [frontends."{{$service.Name}}/{{$partitionId}}"]
          backend = "{{$service.Name}}/{{$partitionId}}"
          [frontends."{{$service.Name}}/{{$partitionId}}".routes.default]
          rule = {{getServiceLabelValue $service "frontend.rule.partition.$partitionId"}}

      {{end}}
    {{end}}
  {{end}}
{{end}}
{{end}}
`
