# Overview

This project creates an Operator using the [Operator Framework](https://github.com/operator-framework), closely following the instructions in their [Getting Started](https://github.com/operator-framework/getting-started) guide. 

It assumes that a [Logstash Helm Chart](https://github.com/helm/charts/tree/master/stable/logstash) has already been deployed in your Cluster, as it indirectly manages resources that it creates.

It's primary purpose is to provide a user-friendly interface for Logstash configuration, with an emphasis on its pattern-matching and filtering capabilities.

## How to Deploy

The Operator can be deployed by creating all the resources under `deploy`:

```
$ kubectl apply -f service_account.yaml
$ kubectl apply -f cluster_role.yaml
$ kubectl apply -f cluster_role_binding.yaml
$ kubectl apply -f operator.yaml 
```

and the custom resource definition (CRD) under `deploy/crds`:

```
$ kubectl apply -f logging.custom_logstashes_crd.yaml
```

Make sure that you deploy all the resources, with the exception of `cluster_role.yaml` and `cluster_role_binding.yaml` since they're not namespace-scoped, in the namespace referenced in `cluster_role_binding.yaml`. This is how the operator communicates with the API, and knows to look for changes across all namespaces.

## Creating a CR

Once the Operator and CRD have been created, you should be able to able to create a custom resource (CR). One such example is given at `deploy/crds/logging.custom_v1alpha1_logstash_cr.yaml`. To elaborate, the general form is:

```
apiVersion: logging.custom/v1alpha1
kind: Logstash
...
spec:
  applications:
  - name: <application-name>
    patterns:
      <pattern-name>: <match-pattern>
      ...
    matchers:
    - <filter-pattern>
    ...
  ...
```

where `<application-name>` represents a unique name for an application deployed in your namespace and should be consistent with the name you assigned when you configured [logging](https://github.com/connexta/grayskull/blob/master/docs/kubernetes/features/logging.md), `<pattern-name>` is the name for a pattern you are creating that you can reference under `spec.matchers`, `<match-pattern>` is a grok pattern, and `<filter-pattern>` in the full pattern you want to match on in your log messages. There are [logstash patterns](https://github.com/elastic/logstash/blob/v1.4.2/patterns/grok-patterns) available by default which you should look at before attempting to create your own, as well as [filter-examples](https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html) that you can use to be more familar with the Logstash filter syntax. The Kibana Dashboard also offers a `Grok Debugger` under `Dev Tools`, which allows you to declare your own patterns and experiment with filters so you can be confident your log messages are being indexed the way you expect.


