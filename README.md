# Istio-Linkerd Service Mesh demo

## About

This repo contains a demo showing the usage of both linkerd and istio in kubernets. It contains a few microservices that will run in the service mesh.

## Requirements

minikube running kubernetes >= v1.8.0
  - minikube addons enable ingress
kubectl
helm
helmfile
Docker >= v17.05.0-ce
namerctl

## Services
There are three services:
1. A words service which generates random words
1. A simon says service which calls the words service
1. A capitalization service which calles either the simon service or the words service

## Istio
### Service Graph
http://192.168.99.100:31352/dotviz
### Zipkin
http://192.168.99.100:30601/zipkin/
