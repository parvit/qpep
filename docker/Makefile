
dist-clean: clean
	docker system prune -a

clean:
	docker container ls -aq | xargs -r -n 1 docker rm -f
	docker image ls -aq | xargs -r -n 1 docker image rm -f

linux-build: clean
	cd build-linux && docker compose up

linux-test:
	cd performance-env && docker compose up
