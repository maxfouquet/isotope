from . import case


def test_snakify() -> None:
    assert case.snakify('snakesOnAPlane') == 'snakes_on_a_plane'
    assert case.snakify('SnakesOnAPlane') == 'snakes_on_a_plane'
    assert case.snakify('snakes_on_a_plane') == 'snakes_on_a_plane'


def test_snakify_dict() -> None:
    expected = {
        'snakes_on_a_plane_a': '',
        'snakes_on_a_plane_b': '',
        'snakes_on_a_plane_c': '',
    }
    actual = case.snakify_dict({
        'snakesOnAPlaneA': '',
        'SnakesOnAPlaneB': '',
        'snakes_on_a_plane_c': '',
    })
    assert expected == actual


def test_snakify_recursive_dict() -> None:
    expected = {
        'snakes_on_a_plane': {
            'snakes_on_a_plane': '',
        },
    }
    actual = case.snakify_dict({
        'snakesOnAPlane': {
            'snakesOnAPlane': '',
        },
    })
    assert expected == actual


def test_snakify_dict_with_list_of_dicts() -> None:
    expected = {
        'snakes_on_a_plane': [{
            'snakes_on_a_plane': '',
        }],
    }
    actual = case.snakify_dict({
        'snakesOnAPlane': [{
            'snakesOnAPlane': '',
        }],
    })
    assert expected == actual
