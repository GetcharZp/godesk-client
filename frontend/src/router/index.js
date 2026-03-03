import { createRouter, createWebHistory } from 'vue-router'
import RemoteControl from "../components/RemoteControl.vue";
import RemoteSession from "../components/RemoteSession.vue";
import DeviceList from "../components/DeviceList.vue";
import SystemSettings from "../components/SystemSettings.vue";

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
            path: '/remote-session/:sessionId?',
            name: 'RemoteSession',
            component: RemoteSession,
            props: true,
        },
        {
            path: '/device-list',
            name: 'DeviceList',
            component: DeviceList,
        },
        {
            path: '/system-settings',
            name: 'SystemSettings',
            component: SystemSettings,
        }
    ]
})

export default router
