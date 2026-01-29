import React from 'react';
import { Outlet } from 'react-router-dom';
import { Box, AppBar, Toolbar, Typography, Drawer, List, ListItem, ListItemIcon, ListItemText, Avatar } from '@mui/material';
import { Dashboard as DashboardIcon, Groups as ClubsIcon, People as PlayersIcon, TableChart as TablesIcon, Assessment as ReportsIcon, Settings as SettingsIcon } from '@mui/icons-material';

const drawerWidth = 240;

export default function Layout() {
  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" noWrap component="div">
            B2B Poker Platform
          </Typography>
          <Box sx={{ flexGrow: 1 }} />
          <Avatar sx={{ bgcolor: 'secondary.main' }}>A</Avatar>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box' },
        }}
      >
        <Toolbar />
        <Box sx={{ overflow: 'auto' }}>
          <List>
            {[
              { text: 'Dashboard', icon: <DashboardIcon />, path: '/' },
              { text: 'Clubs', icon: <ClubsIcon />, path: '/clubs' },
              { text: 'Players', icon: <PlayersIcon />, path: '/players' },
              { text: 'Tables', icon: <TablesIcon />, path: '/tables' },
              { text: 'Reports', icon: <ReportsIcon />, path: '/reports' },
              { text: 'Settings', icon: <SettingsIcon />, path: '/settings' },
            ].map((item) => (
              <ListItem button key={item.text} component="a" href={item.path}>
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.text} />
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  );
}
