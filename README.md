
# Work with Personio without API access

This library helps to deal with personio without official API access. 


## Usage/Examples

```go
    p := personio.NewPersonio(personioBaseURL, personioUser, personioPassword)
	p.LoginToPersonio()
	p.SetWorkingTimes(time.Now(), time.Now().Add(time.Minute*4))
```


## License

[MIT](https://choosealicense.com/licenses/mit/)

