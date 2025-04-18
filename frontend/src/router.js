import {createWebHistory, createRouter} from 'vue-router'

import about from './components/about-app.vue'
import home from './components/home-app.vue'
import user from './components/user-app.vue'
import registration from './components/registration-app.vue'
import task from './components/task-app.vue'
import space from './components/space-app.vue'
import rawtext from './components/rawtext-app.vue'
import dashboard from './components/dashboard-app.vue'
import database from './components/database-app.vue'
import rawdata from './components/rawdata-app.vue'
import problem from './components/problem-app.vue'

/* Вариант с подгружаемыми шаблонами vue (побочный эффект недоступность элементов UI при недоступности uimanager) */
/*
const about = () => import('./components/about-app.vue')
const home = () => import('./components/home-app.vue')
const user = () => import('./components/user-app.vue')
const registration = () => import('./components/registration-app.vue')
const task = () => import('./components/task-app.vue')
const space = () => import('./components/space-app.vue')
const rawtext = () => import('./components/rawtext-app.vue')
const rawdata = () => import('./components/rawdata-app.vue')
const dashbord = () => import('./components/dashboard-app.vue')
const database = () => import('./components/database-app.vue')
const metric = () => import('./components/metric-app.vue')
*/

const routes = [
    {
        path: "/",
        name: "Home",
        component: home,
    },
    {
        path: "/about",
        name: "Справка",
        component: about,
    },
    {
        path: "/user",
        name: "Пользователи",
        component: user,
    },
    {
        path: "/rawdata",
        name: "Данные по метрикам",
        component: rawdata,
    },
    {
        path: "/registration",
        name: "Модули",
        component: registration,
    },
    {
        path: "/task",
        name: "Задачи",
        component: task,
    },
    {
        path: "/space",
        name: "Пространства (клиенты)",
        component: space,
    },
    {
        path: "/rawtext",
        name: "Текстовые данные",
        component: rawtext,
    },
    {
        path: "/dashboard",
        name: "Дашборды",
        component: dashboard,
    },
    {
        path: "/problem",
        name: "Проблемы",
        component: problem,
    },
    {
        path: "/database",
        name: "Состояние БД",
        component: database,
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

export default router;