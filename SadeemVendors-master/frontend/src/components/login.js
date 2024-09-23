import React, { useState,useEffect  } from 'react';
import { useNavigate } from 'react-router-dom';
import '../css/login_module.css';

import axios from 'axios';

function AuthForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [image, setImage] = useState(null);
  const [imageName, setImageName] = useState('');
  const [error, setError] = useState(null);
  const [token, setToken] = useState(null);
  const [isSignUp, setIsSignUp] = useState(false);
  const [fieldErrors, setFieldErrors] = useState({
    email: false,
    password: false,
    name: false,
    phone: false,
  });

  const navigate = useNavigate();
  useEffect(() => {
    localStorage.removeItem('token');
    axios.defaults.headers.common['Authorization'] = '';
    setToken(null);
  }, []);
  const handleSubmit = async (event) => {
    event.preventDefault();
  
    const url = isSignUp ? 'http://localhost:8080/signup' : 'http://localhost:8080/signin';
  
    try {
      const formData = new FormData();
      formData.append('email', email);
      formData.append('password', password);
      formData.append('name', name);
      formData.append('phone', phone);
      if (image) formData.append('img', image);
  
      const response = await fetch(url, {
        method: 'POST',
        body: formData,
      });
  
      const data = await response.json();
  
      if (response.ok) {
        if (isSignUp) {
          handleSignIn();
        } else {
          const token = data.token;
          if (token) {
            localStorage.setItem('token', token);
            axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
            navigate('/vendors');
          } else {
            setError('No token received.');
          }
        }
      } else {
        setFieldErrors({
          email: false,
          password: false,
          name: false,
          phone: false,
        });
  
        if (data.error && typeof data.error === 'object') {
          setFieldErrors({
            email: !!data.error.email,
            password: !!data.error.password,
            name: !!data.error.name,
            phone: !!data.error.phone,
          });
  
          const errorMessage = data.error.email || 
                                data.error.password || 
                                data.error.name || 
                                data.error.phone || 
                                'An unknown error occurred.';
  
          setError(errorMessage);
        } else {
          setError(data.error || 'An unknown error occurred.');
        }
      }
    } catch (error) {
      setError('Failed to process request');
    }
  };
  

  const clearError = () => {
    const errorElement = document.querySelector('.error-message');
    if (errorElement) {
      errorElement.classList.add('hide');
      setTimeout(() => {
        setError(null);
        errorElement.classList.remove('hide');
      }, 500); // Match the transition duration
    }
  };

  const handleSignUp = () => {
    clearError();
    setFieldErrors({
      email: false,
      password: false,
      name: false,
      phone: false,
    });
    document.querySelector('.container').classList.add('right-panel-active');
    setTimeout(() => {
      setIsSignUp(true);
    }, 500);
  };

  const handleSignIn = () => {
    clearError();
    setFieldErrors({
      email: false,
      password: false,
      name: false,
      phone: false,
    });
    document.querySelector('.container').classList.remove('right-panel-active');
    setTimeout(() => {
      setIsSignUp(false);
    }, 500);
  };

  const handleImageChange = (event) => {
    const file = event.target.files[0];
    setImage(file);
    setImageName(file ? file.name : 'No file selected');
  };

  return (
    <div class="login-page">
  <div class="login-component">
    <div className="login-component">
      <div className={`container ${isSignUp ? 'right-panel-active' : ''}`}>
        <div className={`form-container ${isSignUp ? 'sign-up-container' : 'sign-in-container'}`}>
          <form onSubmit={handleSubmit}>
            <h1>{isSignUp ? 'Create Account' : 'Sign In'}</h1>
            {isSignUp && (
              <>
                <input
                  type="text"
                  placeholder="Name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  style={{ borderColor: fieldErrors.name ? 'red' : '' }}
                />
              </>
            )}
            <input
              type="text"
              placeholder="Email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              style={{ borderColor: fieldErrors.email ? 'red' : '' }}
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              style={{ borderColor: fieldErrors.password ? 'red' : '' }}
            />
            {isSignUp && (
              <input
                type="text"
                placeholder="Phone Number"
                value={phone}
                onChange={(event) => setPhone(event.target.value)}
                style={{ borderColor: fieldErrors.phone ? 'red' : '' }}
              />
            )}
            {isSignUp && (
              <div className="file-upload">
                <label htmlFor="imageUpload" className="file-upload-label">
                  Upload Image (optional)
                </label>
                <input
                  type="file"
                  id="imageUpload"
                  onChange={handleImageChange}
                  className="file-upload-input"
                />
                <p className="image-name">{imageName}</p>
              </div>
            )}
            {error && <p className={`error-message ${error ? 'show' : ''}`}>{error}</p>}
            <button type="submit">{isSignUp ? 'Sign Up' : 'Sign In'}</button>
          </form>
        </div>
        <div className="overlay-container">
          <div className="overlay">
            <div className="overlay-panel overlay-left">
              <h1>Welcome Back!</h1>
              <p>Access your account to manage your details.</p>
              <button className="ghost" onClick={handleSignIn}>Sign In</button>
            </div>
            <div className="overlay-panel overlay-right">
              <h1>Join Us!</h1>
              <p>Enter your details to start collaborating with us.</p>
              <button className="ghost" onClick={handleSignUp}>Sign Up</button>
            </div>
          </div>
        </div>
        {!error && token && <p className="success-message">Logged in successfully</p>}
      </div>
    </div>
    </div>
    </div>
  );
}

export default AuthForm;