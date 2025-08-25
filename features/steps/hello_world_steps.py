from behave import given, when, then


@given('I have a test environment')
def step_have_test_environment(context):
    """Set up a basic test environment."""
    context.test_environment = True


@when('I run a simple test')
def step_run_simple_test(context):
    """Run a simple test operation."""
    context.test_result = "passed"


@then('I should see that it passes')
def step_see_test_passes(context):
    """Verify the test passes."""
    assert context.test_result == "passed"


@given('I have the number {number:d}')
def step_have_number(context, number):
    """Store a number in the context."""
    if not hasattr(context, 'numbers'):
        context.numbers = []
    context.numbers.append(number)


@when('I add them together')
def step_add_numbers(context):
    """Add all stored numbers together."""
    context.result = sum(context.numbers)


@then('the result should be {expected:d}')
def step_result_should_be(context, expected):
    """Verify the result matches expected value."""
    assert context.result == expected, f"Expected {expected}, but got {context.result}"
