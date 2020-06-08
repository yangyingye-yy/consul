import Helper from '@ember/component/helper';
import { inject as service } from '@ember/service';

export default Helper.extend({
  env: service('env'),
  compute([name, def = ''], hash) {
    const val = this.env.var(name);
    return val != null ? val : def;
  },
});
