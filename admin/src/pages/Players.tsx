import React from 'react';
import { Box, Typography, TextField, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Button } from '@mui/material';

const players = [
  { id: '1', username: 'PlayerOne', balance: 5000, handsPlayed: 1234, status: 'active' },
  { id: '2', username: 'CardShark', balance: 3200, handsPlayed: 567, status: 'active' },
  { id: '3', username: 'LuckyBet', balance: 150, handsPlayed: 89, status: 'suspended' },
];

export default function Players() {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Players</Typography>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <TextField label="Search" size="small" />
          <Button variant="contained">Add Player</Button>
        </Box>
      </Box>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Username</TableCell>
              <TableCell>Balance</TableCell>
              <TableCell>Hands Played</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {players.map((player) => (
              <TableRow key={player.id}>
                <TableCell>{player.username}</TableCell>
                <TableCell>${player.balance}</TableCell>
                <TableCell>{player.handsPlayed}</TableCell>
                <TableCell>{player.status}</TableCell>
                <TableCell>
                  <Button size="small">Edit</Button>
                  <Button size="small" color="error">Suspend</Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
