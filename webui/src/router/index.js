import {createRouter, createWebHashHistory} from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import ConversationView from '../views/ConversationView.vue'
import UsersView from '../views/UsersView.vue'
import GroupsView from '../views/GroupsView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', redirect: '/conversations'},
		{path: '/login', component: LoginView},
		{path: '/conversations', component: ConversationsView},
		{path: '/conversations/:id', component: ConversationView, props: true},
		{path: '/users', component: UsersView},
		{path: '/groups', component: GroupsView},
	]
})

router.beforeEach((to, from, next) => {
	const hasToken = localStorage.getItem('token');
	if (to.path !== '/login' && !hasToken) {
		next('/login');
	} else {
		next();
	}
});

export default router
