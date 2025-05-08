import { createRouter, createWebHistory } from 'vue-router'
import RemoteControl from "../components/RemoteControl.vue";
import DeviceList from "../components/DeviceList.vue";

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            redirect: '/remote-control',
        },
        {
            path: '/remote-control',
            name: 'RemoteControl',
            component: RemoteControl,
        },
        {
            path: '/device-list',
            name: 'DeviceList',
            component: DeviceList,
        }
    ]
})

export default router
