import { helper } from '@ember/component/helper';

export function serviceOtherGateways([item], hash) {
  switch (true) {
    case item === 'mesh-gateway':
      return ['Terminating', 'Ingress'];
    case item === 'terminating-gateway':
      return ['Mesh', 'Ingress'];
    case item === 'ingress-gateway':
      return ['Terminating', 'Mesh'];
    default:
      return '';
  }
}

export default helper(serviceOtherGateways);
