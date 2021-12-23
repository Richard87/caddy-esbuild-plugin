import React from "react"
import logo from './logo.svg';
import './App.scss';

function App() {
    return (
        <div className="App">
            <header className="App-header">
                <img src={logo} className="App-logo" alt="logo"/>
                <p>
                    Edit <code>src/App.js</code> and save to reload.
                </p>
                <a
                    className="App-link"
                    href="https://reactjs.org"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Learn React with EsBuild
                </a>
                {process.env.REACT_APP_TEST}
                <div className={"alert alert-primary"}><h1>Hello world!</h1></div>
            </header>
        </div>
    );
}

export default App;
