package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	tpl, ok := t.templates[name]
	if !ok {
		return errNotFound
	}
	return tpl.Execute(w, data)
}

func NewPageRenderer() *Template {
	return &Template{
		templates: map[string]*template.Template{
			"AuthTemplate":    template.Must(template.New("AuthTemplate").Parse(loginPageTplSrc)),
			"RefreshTemplate": template.Must(template.New("RefreshTemplate").Parse(refreshPageTmpSrc)),
		},
	}
}

type RefreshPageParams struct {
	Title   string
	Message string
	Emoji   string
	Seconds int
}

const loginPageTplSrc = `
<!DOCTYPE html>
<html>

<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login Page</title>
    <style>
        :root {
            --background-color: #f2f2f2;
            --text-color: #000;
            --input-background-color: #fff;
            --button-background-color: #4caf50;
            --button-text-color: #fff;
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --background-color: #333;
                --text-color: #fff;
                --input-background-color: #444;
                --button-background-color: #6abf69;
                --button-text-color: #000;
            }
        }

        body {
            font-family: Arial, sans-serif;
            background-color: var(--background-color);
            color: var(--text-color);
        }

        .container {
            max-width: 400px;
            margin: 0 auto;
            padding: 40px;
            background-color: var(--input-background-color);
            border-radius: 5px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }

        h2 {
            text-align: center;
            margin-bottom: 30px;
        }

        input[type="text"],
        input[type="password"] {
            width: 100%;
            padding: 10px;
            margin-bottom: 20px;
            border: 1px solid #ccc;
            border-radius: 4px;
            box-sizing: border-box;
            background-color: var(--input-background-color);
            color: var(--text-color);
        }

        input[type="submit"] {
            width: 100%;
            padding: 10px;
            background-color: var(--button-background-color);
            color: var(--button-text-color);
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
        }

        input[type="submit"]:hover {
            background-color: darken(var(--button-background-color), 10%);
        }

        @media (max-width: 600px) {
            .container {
                padding: 20px;
            }

            h2 {
                font-size: 24px;
            }

            input[type="submit"] {
                font-size: 16px;
            }
        }
    </style>
</head>

<body>
    <div class="container">
        <h2>Login</h2>
        <form action="{{.Path}}" method="POST">
            <input type="text" name="username" placeholder="Username" required>
            <input type="password" name="password" placeholder="Password" required>
            <input type="submit" value="Login">
        </form>
    </div>
    <script>
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.setAttribute('data-theme', 'dark');
        }

        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            const newColorScheme = e.matches ? 'dark' : 'light';
            document.documentElement.setAttribute('data-theme', newColorScheme);
        });
    </script>
</body>

</html>
`

const refreshPageTmpSrc = `
<!DOCTYPE html>
<html>
<head>
  <title>{{.Title}}</title>
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
  <h1>{{.Title}}</h1>
  <p>{{.Message}}</p>
  <div class="emoji">{{.Emoji}}</div>
  <p id="timer">Refreshing in <span id="countdown">11</span> seconds...</p>

  <script>
    var countdown = {{.Seconds}};
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
