import Service from '@ember/service';
import { env } from 'consul-ui/env';

export default Service.extend({
  var: function(key) {
    return env(key);
  },
});
