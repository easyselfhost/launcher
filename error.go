package main

const refreshableErrorHTML = `
<!DOCTYPE html>
<html>
<head>
  <title>503 Service Unavailable</title>
  <style>
    :root {
      --background-color: #f2f2f2;
      --text-color: #333;
      --button-background-color: #4caf50;
      --button-text-color: #fff;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --background-color: #333;
        --text-color: #fff;
        --button-background-color: #6abf69;
        --button-text-color: #000;
      }
    }

    body {
      font-family: Arial, sans-serif;
      background-color: var(--background-color);
      color: var(--text-color);
      padding: 20px;
      text-align: center;
    }

    h1 {
      font-size: 36px;
      margin-top: 50px;
    }

    p {
      font-size: 18px;
      margin-top: 20px;
    }

    .emoji {
      font-size: 50px;
      margin-top: 50px;
    }

    #timer {
      font-size: 24px;
      margin-top: 30px;
    }

    input[type="button"] {
      padding: 10px;
      background-color: var(--button-background-color);
      color: var(--button-text-color);
      border: none;
      border-radius: 4px;
      cursor: pointer;
      font-weight: bold;
    }

    input[type="button"]:hover {
      background-color: darken(var(--button-background-color), 10%);
    }
  </style>
</head>
<body>
  <h1>503 Service Unavailable</h1>
  <p>Sorry, the server may not be ready.</p>
  <div class="emoji">😞</div>
  <p id="timer">Refreshing in <span id="countdown">11</span> seconds...</p>

  <script>
    var countdown = 11;
    var countdownElement = document.getElementById("countdown");

    function updateCountdown() {
      countdown--;
      countdownElement.textContent = countdown;

      if (countdown <= 0) {
        location.reload();
      } else {
        setTimeout(updateCountdown, 1000); 
      }
    }

    function toggleDarkMode() {
      var root = document.documentElement;
      var currentTheme = root.getAttribute("data-theme");

      if (currentTheme === "dark") {
        root.setAttribute("data-theme", "light");
      } else {
        root.setAttribute("data-theme", "dark");
      }
    }

    updateCountdown();
  </script>
</body>
</html>
`
