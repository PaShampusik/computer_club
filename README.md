# Аналог системы управления компьютерным клубом

Для запуска программы достаточно клонировать репозиторий, выполнить команды, находясь в папке проекта:

* docker build -t task .
* docker run -v $pwd/tests/input.txt:/app/input.txt -it task

Данные команды выполняют программу для примера в файле задания.

Для запуска теста на события с очередью ожидания выполните следующую команду:

* docker run -v $pwd/tests/test_queue.txt:/app/input.txt -it task

Тест на сортировку имен при принудительном высвобождении клиентов: 

* docker run -v $pwd/tests/test_names.txt:/app/input.txt -it task

Смешанный тест с различными комбинациями событий:

* docker run -v $pwd/tests/test_mixed.txt:/app/input.txt -it task

Тест на пересаживание клиента за разные столы:

* docker run -v $pwd/tests/test_transfer.txt:/app/input.txt -it task
