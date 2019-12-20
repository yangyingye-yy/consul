import Route from '@ember/routing/route';
import { inject as service } from '@ember/service';
import { Promise, hash } from 'rsvp';

import WithNspaceActions from 'consul-ui/mixins/nspace/with-actions';

export default Route.extend(WithNspaceActions, {
  repo: service('repository/nspace/enabled'),
  isCreate: function(params, transition) {
    return transition.targetName.split('.').pop() === 'create';
  },
  model: function(params, transition) {
    const create = this.isCreate(...arguments);
    const dc = this.modelFor('dc').dc.Name;
    // Even though users can see namespaces we aren't giving them a button
    // to get here via the UI, so we also stop user from using the URL to get
    // here, even though theoretically they are allowed to see this page
    return hash({
      permissions: this.repo.authorize(dc, this.modelFor('nspace').nspace.substr(1)),
      isLoading: false,
      create: create,
      dc: dc,
    }).then(model => {
      if (
        typeof (model.permissions || []).find(function(item) {
          return item.Resource === 'operator' && item.Access === 'write' && item.Allow;
        }) !== 'undefined'
      ) {
        return hash({
          ...model,
          item: model.create
            ? Promise.resolve(
                this.repo.create({
                  ACLs: {
                    PolicyDefaults: [],
                    RoleDefaults: [],
                  },
                })
              )
            : this.repo.findBySlug(params.name),
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
