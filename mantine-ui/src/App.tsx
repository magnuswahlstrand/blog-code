import React from 'react';
import './App.css';
import AppShellDemo from "./component/AppShell";
import { NotificationsProvider } from '@mantine/notifications';


function App() {
    return (
        <div className="App">
            <NotificationsProvider>
                <AppShellDemo/>
            </NotificationsProvider>

        </div>
    );
}

export default App;
