import React, { useState, useEffect } from 'react';
import logo from './logo.svg';
import './App.css';
import Button from './components/Button';
import Login from './pages/login';
import Dashboard from './pages/dashboard';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);

  const handleClick = () => {
    alert('Chelsea Just Pooped D:');
  }

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    setIsLoggedIn(!!token);
    setLoading(false);
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', marginTop: '50px' }}>
        <p>Loading...</p>
      </div>
    );
  }

  return (
    <div className="App">
      <Button text="ClickMe" onClick={handleClick} /> 
      {isLoggedIn ? <Dashboard/> : <Login/>}
    </div>
  );
}

export default App;
