all:
	go build -o bin/h2t
run: all
	./bin/h2t
debug: all
	./bin/h2t -debug=true	# @see https://gobyexample.com/command-line-flags
release:
	echo "TODO"
clean:
	rm -rf bin/h2t
