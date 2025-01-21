{
  apiVersion: 'networking.gke.io/v1',
  kind: 'ManagedCertificate',
  metadata: {
    name: 'lightdash-doota-ai',
    namespace: 'default',
  },
  spec: {
    domains: [
      'lightdash.doota.ai',
    ],
  },
}
