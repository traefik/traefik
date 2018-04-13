package servicefabric

const tmpl = `
{{$groupedServiceMap := getServices .Services "backend.group.name"}}
[backends]
    {{range $aggName, $aggServices := $groupedServiceMap }}
      [backends."{{$aggName}}"]
      {{range $service := $aggServices}}
        {{range $partition := $service.Partitions}}
          {{range $instance := $partition.Instances}}
            [backends."{{$aggName}}".servers."{{$service.ID}}-{{$instance.ID}}"]
            url = "{{getDefaultEndpoint $instance}}"
            weight = {{getLabelValue $service "backend.group.weight" "1"}}
          {{end}}
        {{end}}
      {{end}}
    {{end}}
  {{range $service := .Services}}
    {{range $partition := $service.Partitions}}
      {{if eq $partition.ServiceKind "Stateless"}}
      [backends."{{$service.Name}}"]
        [backends."{{$service.Name}}".LoadBalancer]
        {{if hasLabel $service "backend.loadbalancer.method"}}
          method = "{{getLabelValue $service "backend.loadbalancer.method" "" }}"
        {{else}}
          method = "drr"
        {{end}}

        {{if hasLabel $service "backend.healthcheck"}}
          [backends."{{$service.Name}}".healthcheck]
          path = "{{getLabelValue $service "backend.healthcheck" ""}}"
          interval = "{{getLabelValue $service "backend.healthcheck.interval" "10s"}}"
        {{end}}

        {{if hasLabel $service "backend.loadbalancer.stickiness"}}
          [backends."{{$service.Name}}".LoadBalancer.stickiness]
        {{end}}

        {{if hasLabel $service "backend.circuitbreaker"}}
          [backends."{{$service.Name}}".circuitbreaker]
          expression = "{{getLabelValue $service "backend.circuitbreaker" ""}}"
        {{end}}

        {{if hasLabel $service "backend.maxconn.amount"}}
          [backends."{{$service.Name}}".maxconn]
          amount = {{getLabelValue $service "backend.maxconn.amount" ""}}
          {{if hasLabel $service "backend.maxconn.extractorfunc"}}
          extractorfunc = "{{getLabelValue $service "backend.maxconn.extractorfunc" ""}}"
          {{end}}
        {{end}}

        {{range $instance := $partition.Instances}}
          [backends."{{$service.Name}}".servers."{{$instance.ID}}"]
          url = "{{getDefaultEndpoint $instance}}"
          weight = {{getLabelValue $service "backend.weight" "1"}}
        {{end}}
      {{else if eq $partition.ServiceKind "Stateful"}}
        {{range $replica := $partition.Replicas}}
          {{if isPrimary $replica}}

            {{$backendName := getBackendName $service $partition}}
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

    {{if hasLabel $service "frontend.priority"}}
    priority = 100
    {{end}}

    {{range $key, $value := getLabelsWithPrefix $service "frontend.rule"}}
    [frontends."{{$groupName}}".routes."{{$key}}"]
    rule = "{{$value}}"
    {{end}}
{{end}}
{{range $service := .Services}}
  {{if isExposed $service}}
    {{if eq $service.ServiceKind "Stateless"}}

    [frontends."{{$service.Name}}"]
    backend = "{{$service.Name}}"

    {{if hasLabel $service "frontend.passHostHeader"}}
      passHostHeader = {{getLabelValue $service "frontend.passHostHeader"  ""}}
    {{end}}

    {{if hasLabel $service "frontend.whitelistSourceRange"}}
      whitelistSourceRange = {{getLabelValue $service "frontend.whitelistSourceRange"  ""}}
    {{end}}

    {{if hasLabel $service "frontend.priority"}}
      priority = {{getLabelValue $service "frontend.priority" ""}}
    {{end}}

    {{if hasLabel $service "frontend.basicAuth"}}
      basicAuth = {{getLabelValue $service "frontend.basicAuth" ""}}
    {{end}}

    {{if hasLabel $service "frontend.entryPoints"}}
      entryPoints = {{getLabelValue $service "frontend.entryPoints" ""}}
    {{end}}

    {{range $key, $value := getLabelsWithPrefix $service "frontend.rule"}}
    [frontends."{{$service.Name}}".routes."{{$key}}"]
    rule = "{{$value}}"
    {{end}}

    {{else if eq $service.ServiceKind "Stateful"}}
      {{range $partition := $service.Partitions}}
        {{$partitionId := $partition.PartitionInformation.ID}}

        {{ $rule := getLabelValue $service (print "frontend.rule.partition." $partitionId) "" }}
        {{if $rule }}
        [frontends."{{ $service.Name }}/{{ $partitionId }}"]
          backend = "{{ getBackendName $service $partition }}"

          [frontends."{{ $service.Name }}/{{ $partitionId }}".routes.default]
            rule = "{{ $rule }}"
        {{end}}

    {{end}}
  {{end}}
{{end}}
{{end}}
`
