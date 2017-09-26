Planisphere Client Ansible Playbook
===================================

Install and configure the Planishpere client for Linux with Ansible

## Run the playbook to install

    sudo ansible-playbook planisphere.yml \
                          -i localhost, \
                          -c local \
                          --extra-vars="planisphere_key=<your user/group key>"

## Variables

Variables to configure the Planisphere client are included in the planisphere.yml file.  Require variables are:

* bin_file: Where to place the planisphere client script
* clone_dir: Where to find the planisphere tools repo clone (ie: path for the client script)
* log_file: Where to log output from the nightly cron
* planisphere_keyfile: location of the keyfile to be created
* planisphere_config: An array of ini_file configs, to setup the planisphere-report config file, and any overrides

The `planisphere_key` variable is also required, but you may prefer to pass it in the  `--extra-vars=` argument to the `ansible-playbook` command.

## Planisphere Config and Overrides

The planisphere-report config and overrides can be setup using the plansiphere_config variable, and are parsed in [Ansible's ini_file module](http://docs.ansible.com/ansible/latest/ini_file_module.html) format. An example:

        planisphere_config:
        - section: config
          name: key-file
          value: "{{ planisphere_keyfile }}"
        - section: overrides
          name: username
          value: chris
        - section: overrides
          name: serial
          value: cheerios
