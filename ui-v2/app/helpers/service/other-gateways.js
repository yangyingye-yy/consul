import { helper } from '@ember/component/helper';

export function serviceOtherGateways([item], hash) {
  switch (true) {
    case item === 'mesh-gateway':
      return ['terminating', 'ingress'];
    case item === 'terminating-gateway':
      return ['mesh', 'ingress'];
    case item === 'ingress-gateway':
      return ['terminating', 'mesh'];
    default:
      return '';
  }
}

export default helper(serviceOtherGateways);
