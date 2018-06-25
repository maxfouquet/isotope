from typing import Dict

from . import dicts


def test_combine_works() -> None:
    expected = {
        'a': '1',
        'b': '',
        'c': '',
        'd': '',
    }

    a = {'a': ''}
    b = {'b': ''}
    c = {'c': ''}
    d = {'d': ''}
    a1 = {'a': '1'}
    actual = dicts.combine(a, b, c, d, a1)

    assert expected == actual
