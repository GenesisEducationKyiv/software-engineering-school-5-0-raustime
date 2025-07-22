# Testing Guide

## Юніт тести

make docker-test-unit

## Інтеграційні тести

make docker-integration-test

## E2E тести

make docker-e2e-test

## Тестування архітектури

go test -v -timeout 2m ./docs/Architecture/tests/
