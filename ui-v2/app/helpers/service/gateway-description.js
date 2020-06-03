import { helper } from '@ember/component/helper';

export function serviceGatewayDescription([item], hash) {
  switch (true) {
    case item === 'mesh-gateway':
      return 'Mesh gateways enable routing of Connect traffic between different Consul datacenters.';
    case item === 'terminating-gateway':
      return 'Terminating gateways enable connections from services in the Consul service mesh to services outside the mesh.';
    case item === 'ingress-gateway':
      return 'Ingress gateways enable ingress traffic from services outside the Consul service mesh to services inside the Consul service mesh.';
    default:
      return '';
  }
}

export default helper(serviceGatewayDescription);
