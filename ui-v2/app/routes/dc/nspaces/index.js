import Route from '@ember/routing/route';
import { inject as service } from '@ember/service';
import { hash } from 'rsvp';

import WithNspaceActions from 'consul-ui/mixins/nspace/with-actions';
export default Route.extend(WithNspaceActions, {
  repo: service('repository/nspace/enabled'),
  queryParams: {
    s: {
      as: 'filter',
      replace: true,
    },
  },
  model: function(params) {
    // Even though users can see namespaces we aren't giving them a button
    // to get here via the UI, so we also stop user from using the URL to get
    // here, even though theoretically they are allowed to see this page
    return hash({
      permissions: this.repo.authorize(
        this.modelFor('dc').dc.Name,
        this.modelFor('nspace').nspace.substr(1)
      ),
      isLoading: false,
    }).then(model => {
      if (
        typeof (model.permissions || []).find(function(item) {
          return item.Resource === 'operator' && item.Access === 'write' && item.Allow;
        }) !== 'undefined'
      ) {
        return hash({
          ...model,
          items: this.repo.findAll(),
        });
      }
      const e = new Error();
      e.errors = [
        {
          detail: 'Unauthorized',
          status: '401',
        },
      ];
      throw e;
    });
  },
  setupController: function(controller, model) {
    controller.setProperties(model);
  },
});
