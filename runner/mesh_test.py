from unittest import mock

from . import mesh


def test_context_should_call_functions():
    set_up = mock.MagicMock()
    tear_down = mock.MagicMock()
    ingress_url = 'http://example.com'
    get_ingress_url = mock.MagicMock(return_value=ingress_url)
    env = mesh.Environment('', set_up, tear_down, get_ingress_url)
    with env.context() as url:
        set_up.assert_called_once_with()
        get_ingress_url.assert_called_once_with()
        assert url == ingress_url
    tear_down.assert_called_once_with()


def test_context_should_not_call_tear_down_if_should_tear_down_is_false():
    set_up = mock.MagicMock()
    tear_down = mock.MagicMock()
    ingress_url = 'http://example.com'
    get_ingress_url = mock.MagicMock(return_value=ingress_url)
    env = mesh.Environment('', set_up, tear_down, get_ingress_url)
    with env.context(should_tear_down=False):
        pass
    tear_down.assert_not_called()
