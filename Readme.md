# Common Library
This is a set of common functions that I use across projects. 

## Installation

```bash
go get github.com/xyroscar/common-lib
```

## Quick Start

### Configuration

```go
import "github.com/xyroscar/common-lib/pkg/config"

// Set the app name before you load the config
config.SetAppName("myapp")

// This is to register any modules you might need. You can omit this if you do not want to use any modules. 
config.RegisterModule(&config.SmtpConfig{})

// Load the configuration
cfg := config.GetConfig()
```

### SMTP

After loading the initial configuration you can send an email using the below code:
```go
import "github.com/xyroscar/common-lib/pkg/smtp"

email := &smtp.Email{
    To:      []string{"recipient@example.com"},
    Subject: "Hello World",
    Body:    "This is the email body",
    IsHTML:  false,
}

if err := smtp.Send(email); err != nil {
    log.Fatal(err)
}
```

### Logger
There is a default logger present for cases where the global logger isn't initialized. It can be used as follows: 
```go
import "github.com/xyroscar/common-lib/pkg/logger"

logger.Debug("hello", zap.String("user": "user-1"))
logger.Info("hello", zap.String("user": "user-1"))
logger.Warn("hello", zap.String("user": "not found"))
logger.Error("hello", zap.Error(errors.New("user not found")))
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.