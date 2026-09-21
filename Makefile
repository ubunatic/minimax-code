.PHONY: ⚙️ 🤖
⚙️:
🤖:

test-q1: 🤖  # run tests under Quota-1 enforcement
	harnez exec --quota-1 -- $(MAKE) test
