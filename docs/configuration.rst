###############
 Configuration
###############

``quantum`` is designed to handle multiple configuration inputs to allow for as much flexibility as possible. Each different configuration input participates in a configuration hierarchy, which operates as follows first being highest precedence and last being lowest precedence:

  #. Command line arguments.
  #. Environment variables.
  #. Configuration file settings.
  #. Built in defaults.

Hierarchy
=========

The configuration option hierarchy was specifically designed to allow for operators to deploy quantum however they like and leverage what ever configuration management practices they are accustomed to. This means that an operator can deploy a general configuration template file for all ``quantum`` enabled servers, and then override settings in different data-centers using environment variables, and then further override settings on each host by specifying command line arguments.

Defaults
========

The defaults in ``quantum`` are set for ease of access to the system, and meant for development and setting up a proof of concept. In order to properly run ``quantum`` in a production environment please see the :doc:`page on security <security>`. However that being said most of the defaults can be used safely and operators are encouraged to use the defaults where possible unless a need arises to change them or the option is called out as something to change bellow.

Option Reference
================

The general configuration of ``quantum`` is split up into a few different sections outlined bellow. Every option participates in the hierarchy outlined above, and come with defaults, meaning there are no "required" options to run ``quantum``. However there are key configuration options that should be changed for production deployments, as stated above, and they will be called out in the description of the option.

{% for section in config_options.sections %}
{{ section.name }}
{{ "-" * section.name|length }}

----------------------

{{ section.description }}

{% for option in section.options %}
{{ option.name }}
{{ "^" * option.name|length }}

{{ option.description }}

{{ option.type }}
  {{ option.type_def }}

.. list-table::
   :widths: auto
   :header-rows: 1
   :align: center

   * - Command Line (short|long)
     - Environment Variable
     - Configuration File
     - Default
   * - ``-{{ option.short }}|--{{ option.long }}``
     - ``QUANTUM_{{ option.long | upper | replace('-','_') }}``
     - ``{{ option.long }}``
     - ``{% if option.default != "" %}{{ option.default }}{% else %} {% endif %}``

{% endfor %}
{% endfor %}
