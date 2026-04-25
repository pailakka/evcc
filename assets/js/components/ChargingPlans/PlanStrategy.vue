<template>
	<div class="collapsible-wrapper" :class="{ open: show }">
		<div class="collapsible-content pb-3">
			<div v-if="disabled" class="row mb-4">
				<div class="small text-muted">
					<strong class="text-primary">{{ $t("general.note") }}</strong>
					{{ $t("main.chargingPlan.strategyDisabledDescription") }}
				</div>
			</div>
			<div v-else class="row">
				<div class="col-12 col-sm-6 col-lg-3 mb-3">
					<div class="row">
						<label :for="formId('continuous')" class="col-form-label col-5 col-sm-12">
							{{ $t("main.chargingPlan.optimization.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('continuous')"
								v-model="localContinuous"
								class="form-select"
								@change="updateStrategy"
							>
								<option :value="false">
									{{ $t("main.chargingPlan.optimization.cheapest") }}
								</option>
								<option :value="true">
									{{ $t("main.chargingPlan.optimization.continuous") }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div class="col-sm-6 col-lg-3 mb-3">
					<div class="row">
						<label :for="formId('precondition')" class="col-form-label col-5 col-sm-12">
							{{ $t("main.chargingPlan.precondition.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('precondition')"
								v-model="localPrecondition"
								class="form-select"
								@change="updateStrategy"
							>
								<option :value="0">
									{{ $t("main.chargingPlan.precondition.optionNo") }}
								</option>
								<option
									v-for="opt in preconditionOptions"
									:key="opt.value"
									:value="opt.value"
								>
									{{ opt.name }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div v-if="preconditionEnabled" class="col-sm-6 col-lg-3 mb-3">
					<div class="row">
						<label
							:for="formId('preconditionContribution')"
							class="col-form-label col-5 col-sm-12"
						>
							{{ $t("main.chargingPlan.preconditionContribution.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('preconditionContribution')"
								v-model="localPreconditionContribution"
								class="form-select"
								@change="updateStrategy"
							>
								<option
									v-for="opt in preconditionContributionOptions"
									:key="opt.value"
									:value="opt.value"
								>
									{{ opt.name }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div v-if="preconditionEnabled" class="col-sm-6 col-lg-3 mb-3">
					<div class="row">
						<label
							:for="formId('preconditionSupportMode')"
							class="col-form-label col-5 col-sm-12"
						>
							{{ $t("main.chargingPlan.preconditionSupport.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('preconditionSupportMode')"
								v-model="localPreconditionSupportMode"
								class="form-select"
								@change="updateStrategy"
							>
								<option value="">
									{{ $t("main.chargingPlan.preconditionSupport.optionNo") }}
								</option>
								<option value="keepalive">
									{{
										$t("main.chargingPlan.preconditionSupport.optionKeepAlive")
									}}
								</option>
							</select>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import formatter from "@/mixins/formatter";
import type { PlanStrategy } from "./types";

const normalizedContribution = (value?: number) =>
	value === undefined || Number.isNaN(value) ? 1 : value;

export default defineComponent({
	name: "ChargingPlanStrategy",
	mixins: [formatter],
	props: {
		id: [String, Number],
		strategy: Object as PropType<PlanStrategy>,
		show: Boolean,
		precondition: { type: Number, default: 0 },
		continuous: { type: Boolean, default: false },
		disabled: Boolean,
	},
	emits: ["update"],
	data() {
		return {
			localPrecondition: this.precondition,
			localContinuous: this.continuous,
			localPreconditionContribution: normalizedContribution(
				this.strategy?.preconditionContribution
			),
			localPreconditionSupportMode: this.strategy?.preconditionSupportMode || "",
		};
	},
	computed: {
		preconditionEnabled() {
			return this.localPrecondition > 0;
		},
		preconditionOptions() {
			const HOUR = 60 * 60;
			const QUARTER_HOUR = 0.25 * HOUR;
			const HALF_HOUR = 0.5 * HOUR;
			const ONE_HOUR = 1 * HOUR;
			const TWO_HOURS = 2 * HOUR;
			const EVERYTHING = 7 * 24 * HOUR;

			const options = [QUARTER_HOUR, HALF_HOUR, ONE_HOUR, TWO_HOURS, EVERYTHING];

			// support custom values (via API)
			if (this.localPrecondition && !options.includes(this.localPrecondition)) {
				options.push(this.localPrecondition);
			}

			return options.map((s) => ({
				value: s,
				name:
					s === EVERYTHING
						? this.$t("main.chargingPlan.precondition.optionAll")
						: this.fmtDurationLong(s),
			}));
		},
		preconditionContributionOptions() {
			return [0, 0.25, 0.5, 0.75, 1].map((value) => ({
				value,
				name: this.$t("main.chargingPlan.preconditionContribution.option", {
					value: this.fmtPercentage(value * 100),
				}),
			}));
		},
	},
	watch: {
		precondition: {
			handler(newValue: number) {
				// Only update if value actually changed from external source
				if (newValue !== this.localPrecondition) {
					this.localPrecondition = newValue;
				}
			},
			immediate: true,
		},
		continuous: {
			handler(newValue: boolean) {
				// Only update if value actually changed from external source
				if (newValue !== this.localContinuous) {
					this.localContinuous = newValue;
				}
			},
			immediate: true,
		},
		strategy: {
			deep: true,
			handler(newValue: PlanStrategy | undefined) {
				const contribution = normalizedContribution(newValue?.preconditionContribution);
				if (contribution !== this.localPreconditionContribution) {
					this.localPreconditionContribution = contribution;
				}

				const supportMode = newValue?.preconditionSupportMode || "";
				if (supportMode !== this.localPreconditionSupportMode) {
					this.localPreconditionSupportMode = supportMode;
				}
			},
			immediate: true,
		},
	},
	methods: {
		formId(name: string) {
			return `chargingplan-${this.id}-${name}`;
		},
		updateStrategy(): void {
			const strategy: PlanStrategy = {
				...(this.strategy || {}),
				continuous: this.localContinuous,
				precondition: this.localPrecondition,
				preconditionContribution: this.localPreconditionContribution,
				preconditionSupportMode: this
					.localPreconditionSupportMode as PlanStrategy["preconditionSupportMode"],
			};
			this.$emit("update", strategy);
		},
	},
});
</script>
