import React, {useState} from 'react';
import {AppShell, Group, Button, Avatar, Burger, Center, Header, MediaQuery, Navbar, Text, useMantineTheme,} from '@mantine/core';
import { showNotification } from '@mantine/notifications';

export default function AppShellDemo() {
    const theme = useMantineTheme();
    const [opened, setOpened] = useState(false);

    return (
        <AppShell
            styles={{
                main: {
                    background: theme.colorScheme === 'dark' ? theme.colors.dark[8] : theme.colors.gray[0],
                },
            }}
            navbarOffsetBreakpoint="sm"
            asideOffsetBreakpoint="sm"
            fixed
            navbar={
                <Navbar p="md" hiddenBreakpoint="sm" hidden={!opened} width={{sm: 200, lg: 300}}>
                    <Center>
                        <Avatar radius="lg" size="xl"
                                src="https://www.gravatar.com/avatar/d706e0820cfb7b8facd8da5588e60b38?d=identicon"/>
                    </Center>
                    <Text size="sm">admin@32b.se</Text>
                </Navbar>
            }
            header={
                <Header height={70} p="md">
                    <div style={{display: 'flex', alignItems: 'center', height: '100%'}}>
                        <MediaQuery largerThan="sm" styles={{display: 'none'}}>
                            <Burger
                                opened={opened}
                                onClick={() => setOpened((o) => !o)}
                                size="sm"
                                color={theme.colors.gray[6]}
                                mr="xl"
                            />
                        </MediaQuery>

                        <Avatar size="md"
                                src="https://www.gravatar.com/avatar/d706e0820cfb7b8facd8da5588e60b38?d=identicon"/>
                        <Text>Application header</Text>
                    </div>
                </Header>
            }
        >
            <Text>

                <Group position="center">
                    <Button
                        variant="outline"
                        onClick={() =>
                            showNotification({
                                title: 'Default notification',
                                message: 'Hey there, your code is awesome!',
                                autoClose: 1000,
                            })
                        }
                    >Show notification</Button>
                    <Button
                        variant="outline"
                        color="red"
                        onClick={() =>
                            showNotification({
                                title: 'Default notification',
                                message: 'Hey there, your code is awesome!',
                                color: 'dark',
                                autoClose: 1000,
                            })
                        }
                    >Show notification2</Button>


                </Group>
            </Text>
        </AppShell>
    );
}
